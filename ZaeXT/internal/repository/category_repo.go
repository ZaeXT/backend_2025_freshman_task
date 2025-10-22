package repository

import (
	"ai-qa-backend/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *model.Category) error
	GetByID(id, userID uint) (*model.Category, error)
	ListByUserID(userID uint) ([]*model.Category, error)
	Update(category *model.Category) error
	DeleteByID(id, userID uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id, userID uint) (*model.Category, error) {
	var category model.Category
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error
	return &category, err
}

func (r *categoryRepository) ListByUserID(userID uint) ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.Where("user_id = ? AND parent_id IS NULL", userID).Preload("Children").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Model(category).Where("id = ? AND user_id = ?", category.ID, category.UserID).Select("Name", "ParentID").Updates(category).Error
}

func (r *categoryRepository) DeleteByID(id, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var idsToDelete []uint

		recursiveQuery := `
            WITH RECURSIVE descendant_ids AS (
                SELECT id FROM categories WHERE id = ? AND user_id = ?
                UNION ALL
                SELECT c.id FROM categories c JOIN descendant_ids d ON c.parent_id = d.id
            ) SELECT id FROM descendant_ids;
        `

		if err := tx.Raw(recursiveQuery, id, userID).Scan(&idsToDelete).Error; err != nil {
			return err
		}

		if len(idsToDelete) == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(&model.Conversation{}).Where("category_id IN ?", idsToDelete).Update("category_id", nil).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Category{}).Error; err != nil {
			return err
		}

		return nil
	})
}
