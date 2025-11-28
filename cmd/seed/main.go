package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"taxifleet/backend/internal/config"
	"taxifleet/backend/internal/database"
	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sirupsen/logrus"
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

	// Initialize repository
	repo := repository.New(db.GetDB())

	// Check if tenant already exists
	tenant, err := repo.GetTenantBySubdomain("gnakpa-transport")
	if err == nil {
		// No error means tenant was found
		logger.Info("Default tenant already exists, skipping seed")
		fmt.Println("Tenant 'Gnakpa Transport' already exists in the database.")
		os.Exit(0)
	}

	// Check if error is "record not found" (expected) or something else
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.WithError(err).Fatal("Failed to check for existing tenant")
	}

	// If we get here, tenant doesn't exist (err is ErrRecordNotFound)
	logger.Info("Tenant not found, creating default tenant and seed data...")

	// Create tenant
	tenant = &repository.Tenant{
		Name:      "Gnakpa Transport",
		Subdomain: "gnakpa-transport",
		Settings:  "{}", // Valid JSON for JSONB column
	}
	if err := repo.CreateTenant(tenant); err != nil {
		logger.WithError(err).Fatal("Failed to create tenant")
	}
	logger.Info("Created tenant: Gnakpa Transport")

	// Helper function to hash password
	hashPassword := func(password string) string {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
		return string(hash)
	}

	// Create users
	users := []struct {
		Email     string
		Password  string
		Role      string
		FirstName string
		LastName  string
		Phone     string
	}{
		{"admin@taxifleet.ci", "admin123", "admin", "Admin", "Admin", "+1234567890"},
		{"tanguy.gnakpa@gnakpa-transport.com", "owner123", "owner", "Tanguy", "Gnakpa", "+1234567890"},
		{"manager@gnakpa-transport.com", "manager123", "manager", "Gnakpa", "Sister", "+1234567891"},
		{"mechanic@gnakpa-transport.com", "mechanic123", "mechanic", "Just", "Mechanicer", "+1234567892"},
		{"driver@gnakpa-transport.com", "driver123", "driver", "Test", "Driver", "+1234567893"},
	}

	for _, u := range users {
		// Convert role to permission
		userPermission := permissions.GetPermissionForRole(u.Role)

		user := &repository.User{
			TenantID:     tenant.ID,
			Email:        u.Email,
			PasswordHash: hashPassword(u.Password),
			Permission:   userPermission,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			Phone:        u.Phone,
			Active:       true,
		}

		if err := repo.CreateUser(user); err != nil {
			logger.WithError(err).Warnf("Failed to create user %s (may already exist)", u.Email)
		} else {
			logger.Infof("Created user: %s %s (%s, permission: %d)", u.FirstName, u.LastName, u.Role, userPermission)
		}
	}

	// Create test taxis
	taxis := []struct {
		LicensePlate string
		Model        string
		Year         int
		Color        string
		VIN          string
		Status       string
	}{
		{"ABC-123", "Toyota Camry", 2020, "White", "VIN123456789", "active"},
		{"XYZ-789", "Honda Accord", 2021, "Black", "VIN987654321", "active"},
		{"DEF-456", "Nissan Altima", 2019, "Silver", "VIN456789123", "maintenance"},
	}

	for _, t := range taxis {
		taxi := &repository.Taxi{
			TenantID:     tenant.ID,
			LicensePlate: t.LicensePlate,
			Model:        t.Model,
			Year:         t.Year,
			Color:        t.Color,
			VIN:          t.VIN,
			Status:       t.Status,
		}

		if err := repo.CreateTaxi(taxi); err != nil {
			logger.WithError(err).Warnf("Failed to create taxi %s (may already exist)", t.LicensePlate)
		} else {
			logger.Infof("Created taxi: %s", t.LicensePlate)
		}
	}

	logger.Info("Seed data created successfully!")
	fmt.Println("\n=== Seed Data Created ===")
	fmt.Println("Tenant: Gnakpa Transport")
	fmt.Println("\nUsers:")
	fmt.Println("  Owner:    tanguy.gnakpa@gnakpa-transport.com / owner123")
	fmt.Println("  Manager:  manager@gnakpa-transport.com / manager123")
	fmt.Println("  Mechanic: mechanic@gnakpa-transport.com / mechanic123")
	fmt.Println("  Driver:   driver@gnakpa-transport.com / driver123")
	fmt.Println("\nTest Taxis:")
	fmt.Println("  ABC-123 (Toyota Camry 2020 - Active)")
	fmt.Println("  XYZ-789 (Honda Accord 2021 - Active)")
	fmt.Println("  DEF-456 (Nissan Altima 2019 - Maintenance)")
}
