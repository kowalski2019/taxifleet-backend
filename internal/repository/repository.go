package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *sqlx.DB) *Repository {
	// Convert sqlx.DB to GORM by getting the underlying *sql.DB
	sqlDB := db.DB
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		panic("Failed to create GORM instance: " + err.Error())
	}
	return &Repository{db: gormDB}
}

// User methods
func (r *Repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.Preload("Tenant").First(&user, id).Error
	return &user, err
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Preload("Tenant").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByPhone(phone string) (*User, error) {
	var user User
	err := r.db.Preload("Tenant").Where("phone = ?", phone).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByPhoneNumber(phoneNumber string) (*User, error) {
	// Search for phone numbers that match the number part (without country code)
	// Format in DB is "+XX YYYYYYYY", so we search for " YYYYYYYY"
	var user User
	err := r.db.Preload("Tenant").Where("phone LIKE ?", "% "+phoneNumber).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByEmailOrPhone(emailOrPhone string) (*User, error) {
	var user User

	// First try email
	err := r.db.Preload("Tenant").Where("email = ?", emailOrPhone).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// If not found by email, try phone number
	// Extract phone number part (without country code) if format is "+XX YYYYYYYY"
	phoneNumber := emailOrPhone

	// If input contains a space, it might be in format "+XX YYYYYYYY"
	if strings.Contains(emailOrPhone, " ") {
		// Format: "+XX YYYYYYYY" - extract the number part (everything after first space)
		parts := strings.Fields(emailOrPhone)
		if len(parts) > 1 {
			// Get everything after first space and remove all non-digits
			phoneNumber = strings.Join(parts[1:], "")
		}
	}

	// Remove all non-digit characters from phone number for matching
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")

	// Search for phone numbers that match the number part
	// Format in DB is "+XX YYYYYYYY", so we search for " YYYYYYYY" (space + number)
	// Also try exact match in case full format was provided
	err = r.db.Preload("Tenant").Where("phone LIKE ? OR phone = ?", "% "+phoneNumber, emailOrPhone).First(&user).Error
	return &user, err
}

func (r *Repository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *Repository) GetAllUsers() ([]User, error) {
	var users []User
	err := r.db.Preload("Tenant").Find(&users).Error
	return users, err
}

func (r *Repository) GetUsersByTenant(tenantID uint) ([]User, error) {
	var users []User
	err := r.db.Preload("Tenant").Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (r *Repository) DeleteUser(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

// Tenant methods
func (r *Repository) CreateTenant(tenant *Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *Repository) GetTenantByID(id uint) (*Tenant, error) {
	var tenant Tenant
	err := r.db.First(&tenant, id).Error
	return &tenant, err
}

func (r *Repository) GetTenantBySubdomain(subdomain string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.Where("subdomain = ?", subdomain).First(&tenant).Error
	return &tenant, err
}

func (r *Repository) GetAllTenants() ([]Tenant, error) {
	var tenants []Tenant
	err := r.db.Find(&tenants).Error
	return tenants, err
}

func (r *Repository) UpdateTenant(tenant *Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *Repository) DeleteTenant(id uint) error {
	return r.db.Delete(&Tenant{}, id).Error
}

// Taxi methods
func (r *Repository) CreateTaxi(taxi *Taxi) error {
	return r.db.Create(taxi).Error
}

func (r *Repository) GetTaxiByID(id uint) (*Taxi, error) {
	var taxi Taxi
	err := r.db.Preload("AssignedDriver").Preload("Tenant").First(&taxi, id).Error
	return &taxi, err
}

func (r *Repository) GetTaxisByTenant(tenantID uint) ([]Taxi, error) {
	var taxis []Taxi
	err := r.db.Preload("AssignedDriver").Where("tenant_id = ?", tenantID).Find(&taxis).Error
	return taxis, err
}

func (r *Repository) UpdateTaxi(taxi *Taxi) error {
	return r.db.Save(taxi).Error
}

func (r *Repository) DeleteTaxi(id uint) error {
	return r.db.Delete(&Taxi{}, id).Error
}

// WeeklyReport methods
func (r *Repository) CreateReport(report *WeeklyReport) error {
	return r.db.Create(report).Error
}

func (r *Repository) GetReportByID(id uint) (*WeeklyReport, error) {
	var report WeeklyReport
	err := r.db.Preload("Taxi").Preload("Driver").Preload("ApprovedBy").Preload("Expenses").First(&report, id).Error
	return &report, err
}

func (r *Repository) GetReportsByTenant(tenantID uint) ([]WeeklyReport, error) {
	var reports []WeeklyReport
	err := r.db.Preload("Taxi").Preload("Driver").Where("tenant_id = ?", tenantID).Order("week_start_date DESC").Find(&reports).Error
	return reports, err
}

func (r *Repository) GetReportsByDriver(driverID uint) ([]WeeklyReport, error) {
	var reports []WeeklyReport
	err := r.db.Preload("Taxi").Where("driver_id = ?", driverID).Order("week_start_date DESC").Find(&reports).Error
	return reports, err
}

func (r *Repository) UpdateReport(report *WeeklyReport) error {
	return r.db.Save(report).Error
}

func (r *Repository) DeleteReport(id uint) error {
	return r.db.Delete(&WeeklyReport{}, id).Error
}

// Expense methods
func (r *Repository) CreateExpense(expense *Expense) error {
	return r.db.Create(expense).Error
}

func (r *Repository) GetExpenseByID(id uint) (*Expense, error) {
	var expense Expense
	err := r.db.Preload("Taxi").Preload("CreatedBy").Preload("Report").First(&expense, id).Error
	return &expense, err
}

func (r *Repository) GetExpensesByTenant(tenantID uint) ([]Expense, error) {
	var expenses []Expense
	err := r.db.Preload("Taxi").Preload("CreatedBy").Where("tenant_id = ?", tenantID).Order("date DESC").Find(&expenses).Error
	return expenses, err
}

func (r *Repository) GetExpensesByReport(reportID uint) ([]Expense, error) {
	var expenses []Expense
	err := r.db.Preload("Taxi").Preload("CreatedBy").Where("report_id = ?", reportID).Find(&expenses).Error
	return expenses, err
}

func (r *Repository) UpdateExpense(expense *Expense) error {
	return r.db.Save(expense).Error
}

func (r *Repository) DeleteExpense(id uint) error {
	return r.db.Delete(&Expense{}, id).Error
}

// BankDeposit methods
func (r *Repository) CreateDeposit(deposit *BankDeposit) error {
	return r.db.Create(deposit).Error
}

func (r *Repository) GetDepositByID(id uint) (*BankDeposit, error) {
	var deposit BankDeposit
	err := r.db.First(&deposit, id).Error
	return &deposit, err
}

func (r *Repository) GetDepositsByTenant(tenantID uint) ([]BankDeposit, error) {
	var deposits []BankDeposit
	err := r.db.Where("tenant_id = ?", tenantID).Order("deposit_date DESC").Find(&deposits).Error
	return deposits, err
}

func (r *Repository) UpdateDeposit(deposit *BankDeposit) error {
	return r.db.Save(deposit).Error
}

func (r *Repository) DeleteDeposit(id uint) error {
	return r.db.Delete(&BankDeposit{}, id).Error
}

// Session methods
func (r *Repository) CreateSession(session *Session) error {
	return r.db.Create(session).Error
}

func (r *Repository) GetSessionByToken(token string) (*Session, error) {
	var session Session
	err := r.db.Preload("User").Where("token = ? AND expires_at > NOW()", token).First(&session).Error
	return &session, err
}

func (r *Repository) DeleteSession(token string) error {
	return r.db.Where("token = ?", token).Delete(&Session{}).Error
}

func (r *Repository) DeleteUserSessions(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&Session{}).Error
}
