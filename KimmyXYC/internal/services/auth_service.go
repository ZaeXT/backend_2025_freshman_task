package services

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"AIBackend/internal/models"
	"AIBackend/pkg/auth"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) Register(email, password, role string) (*models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", errors.New("email and password required")
	}
	if role == "" {
		role = "free"
	}
	var existing models.User
	if err := s.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, "", errors.New("email already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{Email: email, PasswordHash: string(hash), Role: role}
	if err := s.DB.Create(user).Error; err != nil {
		return nil, "", err
	}
	token, err := auth.CreateToken(user.ID, user.Email, user.Role, 24*time.Hour)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("invalid credentials")
		}
		return nil, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}
	token, err := auth.CreateToken(user.ID, user.Email, user.Role, 24*time.Hour)
	if err != nil {
		return nil, "", err
	}
	return &user, token, nil
}
