package handlers

import (
	"net/http"
	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	permission, _ := c.Get("permission")

	// Check if user has permission to view dashboard (admin, owner, manager only)
	// Mechanics and drivers should not have access to financial data
	userPerm := permission.(int)

	// Admin has all permissions
	if userPerm == -1 || userPerm == permissions.PermissionAdmin {
		// Allow access
	} else {
		// Check if user has financial permissions (VIEW_DEPOSITS or VIEW_EXPENSES)
		// Only owners and managers have these permissions, not mechanics or drivers
		hasFinancialAccess := permissions.HasPermission(userPerm, permissions.PermissionViewDeposits) ||
			permissions.HasPermission(userPerm, permissions.PermissionViewExpenses)

		if !hasFinancialAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view dashboard"})
			return
		}
	}

	stats, err := h.service.GetStats(tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
