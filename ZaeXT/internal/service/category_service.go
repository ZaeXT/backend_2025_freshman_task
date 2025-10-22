package service

import (
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/repository"
)

type CategoryService interface {
	Create(userID uint, name string, parentID *uint) (*model.Category, error)
	List(userID uint) ([]*model.Category, error)
	Update(id, userID uint, name string, parentID *uint) error
	Delete(id, userID uint) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

func (s *categoryService) Create(userID uint, name string, parentID *uint) (*model.Category, error) {
	category := &model.Category{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	}
	err := s.categoryRepo.Create(category)
	return category, err
}

func (s *categoryService) List(userID uint) ([]*model.Category, error) {
	return s.categoryRepo.ListByUserID(userID)
}

func (s *categoryService) Update(id, userID uint, name string, parentID *uint) error {
	category, err := s.categoryRepo.GetByID(id, userID)
	if err != nil {
		return err
	}
	category.Name = name
	category.ParentID = parentID
	return s.categoryRepo.Update(category)
}

func (s *categoryService) Delete(id, userID uint) error {
	return s.categoryRepo.DeleteByID(id, userID)
}
