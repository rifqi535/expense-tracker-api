package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"gorm.io/gorm"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}
