package repository

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Subdomain string         `gorm:"uniqueIndex;not null" json:"subdomain"`
	Logo      string         `json:"logo"`
	Settings  string         `gorm:"type:jsonb;default:'{}'" json:"settings"` // JSON string, stored as JSONB
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents a system user
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Permission   int            `gorm:"not null;default:3" json:"permission"` // Integer permission mask
	FirstName    string         `gorm:"not null" json:"first_name"`
	LastName     string         `gorm:"not null" json:"last_name"`
	Phone        string         `json:"phone"`
	Active       bool           `gorm:"default:true" json:"active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Taxi represents a taxi vehicle
type Taxi struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TenantID         uint           `gorm:"not null;index" json:"tenant_id"`
	LicensePlate     string         `gorm:"not null" json:"license_plate"`
	Model            string         `json:"model"`
	Year             int            `json:"year"`
	Color            string         `json:"color"`
	VIN              string         `json:"vin"`
	Status           string         `gorm:"default:'active'" json:"status"` // active, maintenance, inactive
	AssignedDriverID *uint          `json:"assigned_driver_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant         Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	AssignedDriver *User  `gorm:"foreignKey:AssignedDriverID" json:"driver,omitempty"`
}

// WeeklyReport represents a driver's weekly report
type WeeklyReport struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	TaxiID        uint           `gorm:"not null;index" json:"taxi_id"`
	DriverID      uint           `gorm:"not null;index" json:"driver_id"`
	WeekStartDate time.Time      `gorm:"not null" json:"week_start_date"`
	Earnings      float64        `gorm:"not null;default:0" json:"earnings"`
	TotalExpenses float64        `gorm:"default:0" json:"total_expenses"`
	Status        string         `gorm:"default:'draft'" json:"status"` // draft, submitted, approved, rejected
	Notes         string         `gorm:"type:text" json:"notes"`
	SubmittedAt   *time.Time     `json:"submitted_at"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	ApprovedByID  *uint          `json:"approved_by_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant     Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Taxi       Taxi      `gorm:"foreignKey:TaxiID" json:"taxi,omitempty"`
	Driver     User      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	ApprovedBy *User     `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	Expenses   []Expense `gorm:"foreignKey:ReportID" json:"expenses,omitempty"`
}

// Expense represents an expense entry
type Expense struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	ReportID    *uint          `gorm:"index" json:"report_id"` // Optional: can be standalone or part of report
	TaxiID      *uint          `gorm:"index" json:"taxi_id"`
	Category    string         `gorm:"not null" json:"category"` // fuel, maintenance, insurance, repair, cleaning, other
	Amount      float64        `gorm:"not null" json:"amount"`
	Reason      string         `gorm:"type:text" json:"reason"`
	ReceiptURL  string         `json:"receipt_url"`
	Date        time.Time      `gorm:"not null" json:"date"`
	CreatedByID uint           `gorm:"not null" json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant    Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Report    *WeeklyReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	Taxi      *Taxi         `gorm:"foreignKey:TaxiID" json:"taxi,omitempty"`
	CreatedBy User          `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

// BankDeposit represents a bank deposit record
type BankDeposit struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Amount      float64        `gorm:"not null" json:"amount"`
	DepositDate time.Time      `gorm:"not null" json:"deposit_date"`
	PeriodStart time.Time      `gorm:"not null" json:"period_start"`
	PeriodEnd   time.Time      `gorm:"not null" json:"period_end"`
	BankAccount string         `json:"bank_account"`
	ProofURL    string         `json:"proof_url"`
	Notes       string         `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// MaintenanceLog represents a maintenance record
type MaintenanceLog struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	TaxiID      uint           `gorm:"not null;index" json:"taxi_id"`
	Description string         `gorm:"type:text" json:"description"`
	Cost        float64        `json:"cost"`
	Date        time.Time      `gorm:"not null" json:"date"`
	MechanicID  *uint          `json:"mechanic_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Tenant   Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Taxi     Taxi   `gorm:"foreignKey:TaxiID" json:"taxi,omitempty"`
	Mechanic *User  `gorm:"foreignKey:MechanicID" json:"mechanic,omitempty"`
}
