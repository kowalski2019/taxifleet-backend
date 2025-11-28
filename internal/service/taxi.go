package service

import (
	"errors"
	"taxifleet/backend/internal/repository"
)

type TaxiService struct {
	repo *repository.Repository
}

func NewTaxiService(repo *repository.Repository) *TaxiService {
	return &TaxiService{repo: repo}
}

type CreateTaxiRequest struct {
	LicensePlate    string `json:"license_plate" binding:"required"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	Color           string `json:"color"`
	VIN             string `json:"vin"`
	Status          string `json:"status"`
	AssignedDriverID *uint  `json:"assigned_driver_id"`
}

type UpdateTaxiRequest struct {
	LicensePlate    string `json:"license_plate"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	Color           string `json:"color"`
	VIN             string `json:"vin"`
	Status          string `json:"status"`
	AssignedDriverID *uint  `json:"assigned_driver_id"`
}

func (s *TaxiService) Create(tenantID uint, req CreateTaxiRequest) (*repository.Taxi, error) {
	taxi := &repository.Taxi{
		TenantID:        tenantID,
		LicensePlate:    req.LicensePlate,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		VIN:             req.VIN,
		Status:          req.Status,
		AssignedDriverID: req.AssignedDriverID,
	}

	if taxi.Status == "" {
		taxi.Status = "active"
	}

	if err := s.repo.CreateTaxi(taxi); err != nil {
		return nil, err
	}

	return s.repo.GetTaxiByID(taxi.ID)
}

func (s *TaxiService) GetByID(id uint, tenantID uint) (*repository.Taxi, error) {
	taxi, err := s.repo.GetTaxiByID(id)
	if err != nil {
		return nil, err
	}

	if taxi.TenantID != tenantID {
		return nil, errors.New("taxi not found")
	}

	return taxi, nil
}

func (s *TaxiService) List(tenantID uint) ([]repository.Taxi, error) {
	return s.repo.GetTaxisByTenant(tenantID)
}

func (s *TaxiService) Update(id uint, tenantID uint, req UpdateTaxiRequest) (*repository.Taxi, error) {
	taxi, err := s.repo.GetTaxiByID(id)
	if err != nil {
		return nil, err
	}

	if taxi.TenantID != tenantID {
		return nil, errors.New("taxi not found")
	}

	if req.LicensePlate != "" {
		taxi.LicensePlate = req.LicensePlate
	}
	if req.Model != "" {
		taxi.Model = req.Model
	}
	if req.Year != 0 {
		taxi.Year = req.Year
	}
	if req.Color != "" {
		taxi.Color = req.Color
	}
	if req.VIN != "" {
		taxi.VIN = req.VIN
	}
	if req.Status != "" {
		taxi.Status = req.Status
	}
	if req.AssignedDriverID != nil {
		taxi.AssignedDriverID = req.AssignedDriverID
	}

	if err := s.repo.UpdateTaxi(taxi); err != nil {
		return nil, err
	}

	return s.repo.GetTaxiByID(taxi.ID)
}

func (s *TaxiService) Delete(id uint, tenantID uint) error {
	taxi, err := s.repo.GetTaxiByID(id)
	if err != nil {
		return err
	}

	if taxi.TenantID != tenantID {
		return errors.New("taxi not found")
	}

	return s.repo.DeleteTaxi(id)
}

