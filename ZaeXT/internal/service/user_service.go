package service

import (
	"ai-qa-backend/internal/configs"
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/pkg/hash"
	"ai-qa-backend/internal/pkg/jwt"
	"ai-qa-backend/internal/repository"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

type UserService interface {
	Register(username, password string) error
	Login(username, password string) (string, error)
	GetUserByID(id uint) (*model.User, error)
	UpdateUserMemory(id uint, memoryInfo string) error
}

type userService struct {
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	jwtHelper    *jwt.JWT
	jwtExpiresAt time.Duration
}

func NewUserService(userRepo repository.UserRepository, categoryRepo repository.CategoryRepository) UserService {
	return &userService{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		jwtHelper:    jwt.NewJWT(configs.Conf.JWT.Secret),
		jwtExpiresAt: configs.Conf.JWT.Expiration,
	}
}

func (s *userService) Register(username, password string) error {
	_, err := s.userRepo.GetByUsername(username)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("username already exists")
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Tier:         "free",
	}

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	s.createDefaultCategoriesForUser(user.ID)

	return nil
}

func (s *userService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}

	if !hash.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid username or password")
	}

	token, err := s.jwtHelper.GenerateToken(user.ID, user.Username, user.Tier, s.jwtExpiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) UpdateUserMemory(id uint, memoryInfo string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	user.MemoryInfo = memoryInfo
	return s.userRepo.Update(user)
}

func (s *userService) createDefaultCategoriesForUser(userID uint) {
	defaultCategories := []string{"工作学习", "个人生活", "兴趣爱好"}
	for _, name := range defaultCategories {
		category := &model.Category{
			UserID: userID,
			Name:   name,
		}
		err := s.categoryRepo.Create(category)
		if err != nil {
			log.Printf("Failed to create default category '%s' for user %d: %v\n", name, userID, err)
		}
	}
}
