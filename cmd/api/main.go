package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taxifleet/backend/internal/config"
	"taxifleet/backend/internal/database"
	"taxifleet/backend/internal/handlers"
	"taxifleet/backend/internal/middleware"
	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/repository"
	"taxifleet/backend/internal/service"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set Gin mode based on environment
	if cfg.Server.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.WithFields(logrus.Fields{
		"environment": cfg.Server.Environment,
		"version":     cfg.Server.Version,
		"port":        cfg.Server.Port,
	}).Info("Starting TaxiFleet API server")

	// Initialize database
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.WithError(err).Error("Failed to close database connection")
		}
	}()

	// Run database migrations
	if err := db.Migrate(); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}

	// Note: Seed data should be created using: go run cmd/seed/main.go
	// This ensures proper bcrypt password hashing compatible with Go's bcrypt library

	// Initialize repository
	repo := repository.New(db.GetDB())

	// Initialize permissions from config
	permissions.SetPermissionMasks(
		cfg.Permissions.Admin,
		cfg.Permissions.Owner,
		cfg.Permissions.Manager,
		cfg.Permissions.Mechanic,
		cfg.Permissions.Driver,
	)

	// Initialize services
	authService := service.NewAuthService(repo, cfg)
	taxiService := service.NewTaxiService(repo)
	reportService := service.NewReportService(repo)
	depositService := service.NewDepositService(repo)
	expenseService := service.NewExpenseService(repo)
	dashboardService := service.NewDashboardService(repo)
	adminService := service.NewAdminService(repo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	taxiHandler := handlers.NewTaxiHandler(taxiService)
	reportHandler := handlers.NewReportHandler(reportService)
	depositHandler := handlers.NewDepositHandler(depositService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	adminHandler := handlers.NewAdminHandler(adminService)

	// Setup router
	router := setupRouter(
		authHandler,
		taxiHandler,
		reportHandler,
		depositHandler,
		expenseHandler,
		dashboardHandler,
		adminHandler,
		authService,
		cfg,
		logger,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown complete")
	}
}

func setupRouter(
	authHandler *handlers.AuthHandler,
	taxiHandler *handlers.TaxiHandler,
	reportHandler *handlers.ReportHandler,
	depositHandler *handlers.DepositHandler,
	expenseHandler *handlers.ExpenseHandler,
	dashboardHandler *handlers.DashboardHandler,
	adminHandler *handlers.AdminHandler,
	authService *service.AuthService,
	cfg *config.Config,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	// Add logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"error":       param.ErrorMessage,
		}).Info("HTTP Request")
		return ""
	}))

	// Add recovery middleware
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			// Registration removed - only admins can create users
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", middleware.Auth(authService, logger), authHandler.Logout)
			auth.GET("/me", middleware.Auth(authService, logger), authHandler.Me)
			auth.PUT("/profile", middleware.Auth(authService, logger), authHandler.UpdateProfile)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth(authService, logger))
		{
			// Dashboard
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.GetStats)
			}

			// Taxis
			taxis := protected.Group("/taxis")
			{
				taxis.GET("", taxiHandler.List)
				taxis.POST("", taxiHandler.Create)
				taxis.GET("/:id", taxiHandler.Get)
				taxis.PUT("/:id", taxiHandler.Update)
				taxis.DELETE("/:id", taxiHandler.Delete)
			}

			// Reports
			reports := protected.Group("/reports")
			{
				reports.GET("", reportHandler.List)
				reports.POST("", reportHandler.Create)
				reports.GET("/:id", reportHandler.Get)
				reports.PUT("/:id", reportHandler.Update)
				reports.DELETE("/:id", reportHandler.Delete)
				reports.POST("/:id/submit", reportHandler.Submit)
				reports.POST("/:id/approve", reportHandler.Approve)
				reports.POST("/:id/reject", reportHandler.Reject)
			}

			// Deposits
			deposits := protected.Group("/deposits")
			{
				deposits.GET("", depositHandler.List)
				deposits.POST("", depositHandler.Create)
				deposits.GET("/:id", depositHandler.Get)
				deposits.PUT("/:id", depositHandler.Update)
				deposits.DELETE("/:id", depositHandler.Delete)
			}

			// Expenses
			expenses := protected.Group("/expenses")
			{
				expenses.GET("", expenseHandler.List)
				expenses.POST("", expenseHandler.Create)
				expenses.GET("/:id", expenseHandler.Get)
				expenses.PUT("/:id", expenseHandler.Update)
				expenses.DELETE("/:id", expenseHandler.Delete)
			}

			// Export
			export := protected.Group("/export")
			{
				export.GET("/reports", reportHandler.Export)
				export.GET("/expenses", expenseHandler.Export)
				export.GET("/deposits", depositHandler.Export)
			}

			// Admin routes (admin only)
			admin := protected.Group("/admin")
			admin.Use(adminHandler.RequireAdmin)
			{
				// Tenant management
				tenants := admin.Group("/tenants")
				{
					tenants.GET("", adminHandler.GetAllTenants)
					tenants.POST("", adminHandler.CreateTenant)
					tenants.GET("/:id", adminHandler.GetTenant)
					tenants.PUT("/:id", adminHandler.UpdateTenant)
					tenants.DELETE("/:id", adminHandler.DeleteTenant)
				}

				// User management
				users := admin.Group("/users")
				{
					users.GET("", adminHandler.GetAllUsers)
					users.POST("", adminHandler.CreateUser)
					users.GET("/tenant/:tenantId", adminHandler.GetUsersByTenant)
					users.GET("/:id", adminHandler.GetUser)
					users.PUT("/:id", adminHandler.UpdateUser)
					users.DELETE("/:id", adminHandler.DeleteUser)
				}
			}
		}
	}

	return router
}
