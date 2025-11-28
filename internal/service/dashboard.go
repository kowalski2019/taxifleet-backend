package service

import (
	"taxifleet/backend/internal/repository"
)

type DashboardService struct {
	repo *repository.Repository
}

func NewDashboardService(repo *repository.Repository) *DashboardService {
	return &DashboardService{repo: repo}
}

type DashboardStats struct {
	TotalTaxis     int     `json:"total_taxis"`
	ActiveDrivers  int     `json:"active_drivers"`
	PendingReports int     `json:"pending_reports"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalExpenses  float64 `json:"total_expenses"`
	NetRevenue     float64 `json:"net_revenue"`
}

func (s *DashboardService) GetStats(tenantID uint) (*DashboardStats, error) {
	// Get all taxis for tenant
	taxis, err := s.repo.GetTaxisByTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Count total taxis
	totalTaxis := len(taxis)

	// Count active drivers (taxis with assigned drivers and status active)
	activeDriversMap := make(map[uint]bool)
	for _, taxi := range taxis {
		if taxi.Status == "active" && taxi.AssignedDriverID != nil {
			activeDriversMap[*taxi.AssignedDriverID] = true
		}
	}
	activeDrivers := len(activeDriversMap)

	// Get all reports for tenant
	reports, err := s.repo.GetReportsByTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Count pending reports (draft status)
	pendingReports := 0
	totalRevenue := 0.0
	for _, report := range reports {
		if report.Status == "draft" {
			pendingReports++
		}
		// Sum earnings from approved reports only
		if report.Status == "approved" {
			totalRevenue += report.Earnings
		}
	}

	// Get all expenses for tenant
	expenses, err := s.repo.GetExpensesByTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Sum total expenses
	totalExpenses := 0.0
	for _, expense := range expenses {
		totalExpenses += expense.Amount
	}

	// Calculate net revenue (total revenue - total expenses)
	netRevenue := totalRevenue - totalExpenses

	return &DashboardStats{
		TotalTaxis:     totalTaxis,
		ActiveDrivers:  activeDrivers,
		PendingReports: pendingReports,
		TotalRevenue:   totalRevenue,
		TotalExpenses:  totalExpenses,
		NetRevenue:     netRevenue,
	}, nil
}
