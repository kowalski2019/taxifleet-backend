package service

import (
	"errors"
	"taxifleet/backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	repo *repository.Repository
}

func NewAdminService(repo *repository.Repository) *AdminService {
	return &AdminService{repo: repo}
}

// Tenant Management
type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
	Logo      string `json:"logo"`
	Settings  string `json:"settings"`
}

type UpdateTenantRequest struct {
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
	Logo      string `json:"logo"`
	Settings  string `json:"settings"`
}

func (s *AdminService) CreateTenant(req CreateTenantRequest) (*repository.Tenant, error) {
	// Check if subdomain already exists
	_, err := s.repo.GetTenantBySubdomain(req.Subdomain)
	if err == nil {
		// Tenant found, subdomain already exists
		return nil, errors.New("subdomain already exists")
	}
	// If error is not "record not found", it's a real error that should be returned
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// err == gorm.ErrRecordNotFound means tenant doesn't exist, which is what we want

	settings := req.Settings
	if settings == "" {
		settings = "{}"
	}

	tenant := &repository.Tenant{
		Name:      req.Name,
		Subdomain: req.Subdomain,
		Logo:      req.Logo,
		Settings:  settings,
	}

	if err := s.repo.CreateTenant(tenant); err != nil {
		return nil, err
	}

	return s.repo.GetTenantByID(tenant.ID)
}

func (s *AdminService) GetAllTenants() ([]repository.Tenant, error) {
	return s.repo.GetAllTenants()
}

func (s *AdminService) GetTenantByID(id uint) (*repository.Tenant, error) {
	return s.repo.GetTenantByID(id)
}

func (s *AdminService) UpdateTenant(id uint, req UpdateTenantRequest) (*repository.Tenant, error) {
	tenant, err := s.repo.GetTenantByID(id)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Subdomain != "" {
		// Check if subdomain is being changed and if new one exists
		if tenant.Subdomain != req.Subdomain {
			_, err := s.repo.GetTenantBySubdomain(req.Subdomain)
			if err == nil {
				// Tenant found, subdomain already exists
				return nil, errors.New("subdomain already exists")
			}
			// If error is not "record not found", it's a real error that should be returned
			if err != gorm.ErrRecordNotFound {
				return nil, err
			}
			// err == gorm.ErrRecordNotFound means tenant doesn't exist, which is what we want
			tenant.Subdomain = req.Subdomain
		}
	}
	if req.Logo != "" {
		tenant.Logo = req.Logo
	}
	if req.Settings != "" {
		tenant.Settings = req.Settings
	}

	if err := s.repo.UpdateTenant(tenant); err != nil {
		return nil, err
	}

	return s.repo.GetTenantByID(tenant.ID)
}

func (s *AdminService) DeleteTenant(id uint) error {
	return s.repo.DeleteTenant(id)
}

// User Management
type CreateUserRequest struct {
	TenantID   uint   `json:"tenant_id" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Permission int    `json:"permission" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Active     bool   `json:"active"`
}

type UpdateUserRequest struct {
	TenantID   uint   `json:"tenant_id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Permission int    `json:"permission"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone" binding:"required"`
	Active     *bool  `json:"active"`
}

func (s *AdminService) CreateUser(req CreateUserRequest) (*repository.User, error) {
	// Verify tenant exists
	_, err := s.repo.GetTenantByID(req.TenantID)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	// Check if email already exists
	_, err = s.repo.GetUserByEmail(req.Email)
	if err == nil {
		// User found, email already exists
		return nil, errors.New("email already exists")
	}
	// If error is not "record not found", it's a real error that should be returned
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// err == gorm.ErrRecordNotFound means user doesn't exist, which is what we want

	// Check if phone number already exists (phone must be unique globally)
	_, err = s.repo.GetUserByPhone(req.Phone)
	if err == nil {
		// User found, phone already exists
		return nil, errors.New("phone number already exists")
	}
	// If error is not "record not found", it's a real error that should be returned
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// err == gorm.ErrRecordNotFound means phone doesn't exist, which is what we want

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	active := req.Active
	if !active {
		active = true // Default to active
	}

	user := &repository.User{
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Permission:   req.Permission,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Active:       active,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return s.repo.GetUserByID(user.ID)
}

func (s *AdminService) GetAllUsers() ([]repository.User, error) {
	return s.repo.GetAllUsers()
}

func (s *AdminService) GetUsersByTenant(tenantID uint) ([]repository.User, error) {
	return s.repo.GetUsersByTenant(tenantID)
}

func (s *AdminService) GetUserByID(id uint) (*repository.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *AdminService) UpdateUser(id uint, req UpdateUserRequest) (*repository.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.TenantID != 0 {
		// Verify tenant exists
		_, err := s.repo.GetTenantByID(req.TenantID)
		if err != nil {
			return nil, errors.New("tenant not found")
		}
		user.TenantID = req.TenantID
	}

	if req.Email != "" {
		// Check if email is being changed and if new one exists
		if user.Email != req.Email {
			_, err = s.repo.GetUserByEmail(req.Email)
			if err == nil {
				// User found, email already exists
				return nil, errors.New("email already exists")
			}
			// If error is not "record not found", it's a real error that should be returned
			if err != gorm.ErrRecordNotFound {
				return nil, err
			}
			// err == gorm.ErrRecordNotFound means user doesn't exist, which is what we want
			user.Email = req.Email
		}
	}

	// Check if phone number is being changed and if new one exists
	if req.Phone != "" && user.Phone != req.Phone {
		_, err = s.repo.GetUserByPhone(req.Phone)
		if err == nil {
			// User found, phone already exists
			return nil, errors.New("phone number already exists")
		}
		// If error is not "record not found", it's a real error that should be returned
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		// err == gorm.ErrRecordNotFound means phone doesn't exist, which is what we want
	}

	if req.Password != "" {
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.PasswordHash = hashedPassword
	}

	if req.Permission != 0 {
		user.Permission = req.Permission
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}

	if req.LastName != "" {
		user.LastName = req.LastName
	}

	// Phone is required, always update it
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return s.repo.GetUserByID(user.ID)
}

func (s *AdminService) DeleteUser(id uint) error {
	return s.repo.DeleteUser(id)
}

// Helper function to hash password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
