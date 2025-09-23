package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rifqi535/expense-tracker-api/internal/models"
	"gorm.io/gorm"
)

type ExpenseRepo struct {
	db *gorm.DB
}

func NewExpenseRepo(db *gorm.DB) *ExpenseRepo {
	return &ExpenseRepo{db: db}
}

// List dengan filter, sort, pagination
func (r *ExpenseRepo) List(
	ctx context.Context,
	userID uuid.UUID,
	categoryID *uuid.UUID,
	startDate, endDate *time.Time,
	limit, offset int,
	sortBy, order string,
) ([]models.Expense, error) {
	var expenses []models.Expense

	// sanity check / default
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if sortBy != "date" && sortBy != "amount" {
		sortBy = "date"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	order = strings.ToUpper(order) // ASC / DESC

	// pilih kolom sort
	sortColumn := "created_at"
	if sortBy == "amount" {
		sortColumn = "amount"
	}

	// build query dengan GORM
	query := r.db.WithContext(ctx).Model(&models.Expense{}).Where("user_id = ?", userID)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// order + pagination
	err := query.Order(fmt.Sprintf("%s %s", sortColumn, order)).
		Limit(limit).
		Offset(offset).
		Find(&expenses).Error

	return expenses, err
}

// ListByUser: shortcut tanpa filter
func (r *ExpenseRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Expense, error) {
	return r.List(ctx, userID, nil, nil, nil, 10, 0, "date", "desc")
}

// Create: tambah expense baru
func (r *ExpenseRepo) Create(ctx context.Context, e *models.Expense) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// Update: ubah expense milik user
func (r *ExpenseRepo) Update(ctx context.Context, userID, id uuid.UUID, title, description string, amount float64, categoryID uuid.UUID) (bool, error) {
	updates := map[string]interface{}{
		"title":       title,
		"description": description,
		"amount":      amount,
		"category_id": categoryID,
		"updated_at":  time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Expense{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)

	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		// tidak ada baris yang berubah → mungkin ID salah atau user bukan pemilik data
		return false, nil
	}

	return true, nil

}

// Delete: hapus expense milik user
func (r *ExpenseRepo) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Expense{})

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
