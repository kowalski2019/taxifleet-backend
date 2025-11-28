package service

import (
	"errors"
	"time"
	"taxifleet/backend/internal/repository"
)

type DepositService struct {
	repo *repository.Repository
}

func NewDepositService(repo *repository.Repository) *DepositService {
	return &DepositService{repo: repo}
}

type CreateDepositRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	DepositDate string  `json:"deposit_date" binding:"required"`
	PeriodStart string  `json:"period_start" binding:"required"`
	PeriodEnd   string  `json:"period_end" binding:"required"`
	BankAccount string  `json:"bank_account"`
	ProofURL    string  `json:"proof_url"`
	Notes       string  `json:"notes"`
}

type UpdateDepositRequest struct {
	Amount      float64 `json:"amount"`
	DepositDate string  `json:"deposit_date"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	BankAccount string  `json:"bank_account"`
	ProofURL    string  `json:"proof_url"`
	Notes       string  `json:"notes"`
}

func (s *DepositService) Create(tenantID uint, req CreateDepositRequest) (*repository.BankDeposit, error) {
	depositDate, _ := time.Parse("2006-01-02", req.DepositDate)
	periodStart, _ := time.Parse("2006-01-02", req.PeriodStart)
	periodEnd, _ := time.Parse("2006-01-02", req.PeriodEnd)

	deposit := &repository.BankDeposit{
		TenantID:    tenantID,
		Amount:      req.Amount,
		DepositDate: depositDate,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		BankAccount: req.BankAccount,
		ProofURL:    req.ProofURL,
		Notes:       req.Notes,
	}

	if err := s.repo.CreateDeposit(deposit); err != nil {
		return nil, err
	}

	return s.repo.GetDepositByID(deposit.ID)
}

func (s *DepositService) GetByID(id uint, tenantID uint) (*repository.BankDeposit, error) {
	deposit, err := s.repo.GetDepositByID(id)
	if err != nil {
		return nil, err
	}

	if deposit.TenantID != tenantID {
		return nil, errors.New("deposit not found")
	}

	return deposit, nil
}

func (s *DepositService) List(tenantID uint) ([]repository.BankDeposit, error) {
	return s.repo.GetDepositsByTenant(tenantID)
}

func (s *DepositService) Update(id uint, tenantID uint, req UpdateDepositRequest) (*repository.BankDeposit, error) {
	deposit, err := s.repo.GetDepositByID(id)
	if err != nil {
		return nil, err
	}

	if deposit.TenantID != tenantID {
		return nil, errors.New("deposit not found")
	}

	if req.Amount != 0 {
		deposit.Amount = req.Amount
	}
	if req.DepositDate != "" {
		depositDate, _ := time.Parse("2006-01-02", req.DepositDate)
		deposit.DepositDate = depositDate
	}
	if req.PeriodStart != "" {
		periodStart, _ := time.Parse("2006-01-02", req.PeriodStart)
		deposit.PeriodStart = periodStart
	}
	if req.PeriodEnd != "" {
		periodEnd, _ := time.Parse("2006-01-02", req.PeriodEnd)
		deposit.PeriodEnd = periodEnd
	}
	if req.BankAccount != "" {
		deposit.BankAccount = req.BankAccount
	}
	if req.ProofURL != "" {
		deposit.ProofURL = req.ProofURL
	}
	if req.Notes != "" {
		deposit.Notes = req.Notes
	}

	if err := s.repo.UpdateDeposit(deposit); err != nil {
		return nil, err
	}

	return s.repo.GetDepositByID(deposit.ID)
}

func (s *DepositService) Delete(id uint, tenantID uint) error {
	deposit, err := s.repo.GetDepositByID(id)
	if err != nil {
		return err
	}

	if deposit.TenantID != tenantID {
		return errors.New("deposit not found")
	}

	return s.repo.DeleteDeposit(id)
}

