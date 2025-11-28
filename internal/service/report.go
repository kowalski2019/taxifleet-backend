package service

import (
	"errors"
	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/repository"
	"time"
)

type ReportService struct {
	repo *repository.Repository
}

func NewReportService(repo *repository.Repository) *ReportService {
	return &ReportService{repo: repo}
}

type CreateReportRequest struct {
	TaxiID        uint      `json:"taxi_id" binding:"required"`
	WeekStartDate time.Time `json:"week_start_date" binding:"required"`
	Earnings      float64   `json:"earnings" binding:"required"`
	Notes         string    `json:"notes"`
}

type UpdateReportRequest struct {
	WeekStartDate time.Time `json:"week_start_date"`
	Earnings      float64   `json:"earnings"`
	Notes         string    `json:"notes"`
}

func (s *ReportService) Create(tenantID uint, driverID uint, req CreateReportRequest) (*repository.WeeklyReport, error) {
	// Verify taxi belongs to tenant
	taxi, err := s.repo.GetTaxiByID(req.TaxiID)
	if err != nil {
		return nil, errors.New("taxi not found")
	}
	if taxi.TenantID != tenantID {
		return nil, errors.New("taxi not found")
	}

	report := &repository.WeeklyReport{
		TenantID:      tenantID,
		TaxiID:        req.TaxiID,
		DriverID:      driverID,
		WeekStartDate: req.WeekStartDate,
		Earnings:      req.Earnings,
		TotalExpenses: 0,
		Status:        "draft",
		Notes:         req.Notes,
	}

	if err := s.repo.CreateReport(report); err != nil {
		return nil, err
	}

	return s.repo.GetReportByID(report.ID)
}

func (s *ReportService) GetByID(id uint, tenantID uint) (*repository.WeeklyReport, error) {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.TenantID != tenantID {
		return nil, errors.New("report not found")
	}

	return report, nil
}

func (s *ReportService) List(tenantID uint, userID uint, permission int) ([]repository.WeeklyReport, error) {
	// Drivers can only see their own reports (only have view/add report permissions)
	if permission == permissions.PermissionDriver {
		return s.repo.GetReportsByDriver(userID)
	}

	// Owners, managers, and others with view permissions see all tenant reports
	return s.repo.GetReportsByTenant(tenantID)
}

func (s *ReportService) Update(id uint, tenantID uint, driverID uint, permission int, req UpdateReportRequest) (*repository.WeeklyReport, error) {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.TenantID != tenantID {
		return nil, errors.New("report not found")
	}

	// Check if user has global edit permission (owner, manager, admin)
	hasEditPermission := permissions.HasPermission(permission, permissions.PermissionEditReports)

	// If user doesn't have global edit permission, they can only edit their own draft reports
	if !hasEditPermission {
		// User must be the creator of the report
		if report.DriverID != driverID {
			return nil, errors.New("unauthorized")
		}
		// Can only edit draft reports
		if report.Status != "draft" {
			return nil, errors.New("can only edit draft reports")
		}
	} else {
		// Users with edit permission can edit any report, but only if it's in draft or submitted status
		if report.Status == "approved" {
			return nil, errors.New("cannot edit approved reports")
		}
	}

	if !req.WeekStartDate.IsZero() {
		report.WeekStartDate = req.WeekStartDate
	}
	if req.Earnings != 0 {
		report.Earnings = req.Earnings
	}
	if req.Notes != "" {
		report.Notes = req.Notes
	}

	// Recalculate total expenses
	expenses, _ := s.repo.GetExpensesByReport(report.ID)
	total := 0.0
	for _, exp := range expenses {
		total += exp.Amount
	}
	report.TotalExpenses = total

	if err := s.repo.UpdateReport(report); err != nil {
		return nil, err
	}

	return s.repo.GetReportByID(report.ID)
}

func (s *ReportService) Submit(id uint, tenantID uint, driverID uint) (*repository.WeeklyReport, error) {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.TenantID != tenantID {
		return nil, errors.New("report not found")
	}

	if report.DriverID != driverID {
		return nil, errors.New("unauthorized")
	}

	if report.Status != "draft" {
		return nil, errors.New("report already submitted")
	}

	now := time.Now()
	report.Status = "submitted"
	report.SubmittedAt = &now

	if err := s.repo.UpdateReport(report); err != nil {
		return nil, err
	}

	return s.repo.GetReportByID(report.ID)
}

func (s *ReportService) Approve(id uint, tenantID uint, approvedByID uint, permission int) (*repository.WeeklyReport, error) {
	// Only owner or admin can approve reports
	if !permissions.HasAnyPermission(permission, permissions.PermissionOwner, permissions.PermissionAdmin) {
		return nil, errors.New("only owner or admin can approve reports")
	}

	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.TenantID != tenantID {
		return nil, errors.New("report not found")
	}

	if report.Status != "submitted" {
		return nil, errors.New("report must be submitted before approval")
	}

	now := time.Now()
	report.Status = "approved"
	report.ApprovedAt = &now
	report.ApprovedByID = &approvedByID

	if err := s.repo.UpdateReport(report); err != nil {
		return nil, err
	}

	return s.repo.GetReportByID(report.ID)
}

func (s *ReportService) Reject(id uint, tenantID uint) (*repository.WeeklyReport, error) {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}

	if report.TenantID != tenantID {
		return nil, errors.New("report not found")
	}

	if report.Status != "submitted" {
		return nil, errors.New("report must be submitted before rejection")
	}

	report.Status = "rejected"

	if err := s.repo.UpdateReport(report); err != nil {
		return nil, err
	}

	return s.repo.GetReportByID(report.ID)
}

func (s *ReportService) Delete(id uint, tenantID uint, driverID uint, permission int) error {
	report, err := s.repo.GetReportByID(id)
	if err != nil {
		return err
	}

	if report.TenantID != tenantID {
		return errors.New("report not found")
	}

	// Owner/admin can delete reports in any status, including approved
	if permissions.HasAnyPermission(permission, permissions.PermissionOwner, permissions.PermissionAdmin) {
		return s.repo.DeleteReport(id)
	}

	// Driver can only delete their own draft reports
	if permission == permissions.PermissionDriver {
		if report.DriverID != driverID {
			return errors.New("unauthorized")
		}
		if report.Status != "draft" {
			return errors.New("can only delete draft reports")
		}
		return s.repo.DeleteReport(id)
	}

	return errors.New("unauthorized")
}
