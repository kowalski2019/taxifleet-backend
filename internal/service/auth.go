package service

import (
	"errors"
	"time"

	"taxifleet/backend/internal/config"
	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewAuthService(repo *repository.Repository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token        string           `json:"token"`
	RefreshToken string           `json:"refresh_token"`
	User         *repository.User `json:"user"`
}

func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		// User found, email already exists
		return nil, errors.New("user with this email already exists")
	}
	// If error is not "record not found", it's a real error that should be returned
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// err == gorm.ErrRecordNotFound means user doesn't exist, which is what we want

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create default tenant for new user (simplified - in production, handle tenant creation separately)
	tenant := &repository.Tenant{
		Name:      req.FirstName + " " + req.LastName,
		Subdomain: generateSubdomain(req.Email),
		Settings:  "{}", // Valid JSON for JSONB column
	}
	if err := s.repo.CreateTenant(tenant); err != nil {
		return nil, err
	}

	// Set default permission (owner if not specified)
	userPermission := permissions.PermissionOwner
	if req.Role != "" {
		userPermission = permissions.GetPermissionForRole(req.Role)
	}

	// Create user
	user := &repository.User{
		TenantID:     tenant.ID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Permission:   userPermission,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Active:       true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// Generate tokens
	token, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &repository.Session{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiration),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Get user
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.New("user account is inactive")
	}

	// Generate tokens
	token, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &repository.Session{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiration),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	// Get session
	session, err := s.repo.GetSessionByToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Get user
	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Generate new access token
	token, _, err := s.generateTokens(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.repo.DeleteSession(refreshToken)
}

func (s *AuthService) ValidateToken(tokenString string) (*repository.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		// Check if token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	user, err := s.repo.GetUserByID(uint(userID))
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is still active
	if !user.Active {
		return nil, errors.New("user account is inactive")
	}

	return user, nil
}

func (s *AuthService) generateTokens(user *repository.User) (string, string, error) {
	// Access token
	accessClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"permission": user.Permission,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(s.cfg.JWT.Expiration).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (longer expiration)
	refreshClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"permission": user.Permission,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(s.cfg.JWT.RefreshExpiration).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func generateSubdomain(email string) string {
	// Simple subdomain generation from email
	// In production, ensure uniqueness
	return email[:len(email)-len("@example.com")]
}
