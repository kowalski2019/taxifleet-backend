package service

import (
	"errors"
	"time"
	"taxifleet/backend/internal/repository"
)

type ExpenseService struct {
	repo *repository.Repository
}

func NewExpenseService(repo *repository.Repository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

type CreateExpenseRequest struct {
	ReportID   *uint   `json:"report_id"`
	TaxiID     *uint   `json:"taxi_id"`
	Category   string  `json:"category" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
	Reason     string  `json:"reason"`
	ReceiptURL string  `json:"receipt_url"`
	Date       string  `json:"date" binding:"required"`
}

type UpdateExpenseRequest struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Reason     string  `json:"reason"`
	ReceiptURL string  `json:"receipt_url"`
	Date       string  `json:"date"`
}

func (s *ExpenseService) Create(tenantID uint, createdByID uint, req CreateExpenseRequest) (*repository.Expense, error) {
	expense := &repository.Expense{
		TenantID:    tenantID,
		ReportID:    req.ReportID,
		TaxiID:      req.TaxiID,
		Category:    req.Category,
		Amount:      req.Amount,
		Reason:      req.Reason,
		ReceiptURL:  req.ReceiptURL,
		CreatedByID: createdByID,
	}

	// Parse date
	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		expense.Date = date
	} else {
		expense.Date = time.Now()
	}

	if err := s.repo.CreateExpense(expense); err != nil {
		return nil, err
	}

	// If expense is part of a report, update report total
	if req.ReportID != nil {
		report, _ := s.repo.GetReportByID(*req.ReportID)
		if report != nil && report.TenantID == tenantID {
			expenses, _ := s.repo.GetExpensesByReport(report.ID)
			total := 0.0
			for _, exp := range expenses {
				total += exp.Amount
			}
			report.TotalExpenses = total
			s.repo.UpdateReport(report)
		}
	}

	return s.repo.GetExpenseByID(expense.ID)
}

func (s *ExpenseService) GetByID(id uint, tenantID uint) (*repository.Expense, error) {
	expense, err := s.repo.GetExpenseByID(id)
	if err != nil {
		return nil, err
	}

	if expense.TenantID != tenantID {
		return nil, errors.New("expense not found")
	}

	return expense, nil
}

func (s *ExpenseService) List(tenantID uint) ([]repository.Expense, error) {
	return s.repo.GetExpensesByTenant(tenantID)
}

func (s *ExpenseService) Update(id uint, tenantID uint, req UpdateExpenseRequest) (*repository.Expense, error) {
	expense, err := s.repo.GetExpenseByID(id)
	if err != nil {
		return nil, err
	}

	if expense.TenantID != tenantID {
		return nil, errors.New("expense not found")
	}

	if req.Category != "" {
		expense.Category = req.Category
	}
	if req.Amount != 0 {
		expense.Amount = req.Amount
	}
	if req.Reason != "" {
		expense.Reason = req.Reason
	}
	if req.ReceiptURL != "" {
		expense.ReceiptURL = req.ReceiptURL
	}
	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		expense.Date = date
	}

	if err := s.repo.UpdateExpense(expense); err != nil {
		return nil, err
	}

	// Update report total if expense is part of a report
	if expense.ReportID != nil {
		report, _ := s.repo.GetReportByID(*expense.ReportID)
		if report != nil {
			expenses, _ := s.repo.GetExpensesByReport(report.ID)
			total := 0.0
			for _, exp := range expenses {
				total += exp.Amount
			}
			report.TotalExpenses = total
			s.repo.UpdateReport(report)
		}
	}

	return s.repo.GetExpenseByID(expense.ID)
}

func (s *ExpenseService) Delete(id uint, tenantID uint) error {
	expense, err := s.repo.GetExpenseByID(id)
	if err != nil {
		return err
	}

	if expense.TenantID != tenantID {
		return errors.New("expense not found")
	}

	reportID := expense.ReportID

	if err := s.repo.DeleteExpense(id); err != nil {
		return err
	}

	// Update report total if expense was part of a report
	if reportID != nil {
		report, _ := s.repo.GetReportByID(*reportID)
		if report != nil {
			expenses, _ := s.repo.GetExpensesByReport(report.ID)
			total := 0.0
			for _, exp := range expenses {
				total += exp.Amount
			}
			report.TotalExpenses = total
			s.repo.UpdateReport(report)
		}
	}

	return nil
}

