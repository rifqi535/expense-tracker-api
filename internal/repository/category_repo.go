package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"gorm.io/gorm"
)

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("title").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil

}

func (r *CategoryRepo) Create(ctx context.Context, c *models.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepo) Update(ctx context.Context, userID, id uuid.UUID, title string) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"title": title,
		})

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *CategoryRepo) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Category{})

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
