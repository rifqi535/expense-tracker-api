package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rifqi535/expense-tracker-api/internal/models"
)

type ExpenseRepo struct {
	db *pgxpool.Pool
}

func NewExpenseRepo(db *pgxpool.Pool) *ExpenseRepo {
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

	// build query dinamis
	query := `
		SELECT id, title, description, amount, category_id, user_id, created_at, updated_at
		FROM expenses
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{userID}
	argPos := 2

	if categoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argPos)
		args = append(args, *categoryID)
		argPos++
	}
	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *startDate)
		argPos++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *endDate)
		argPos++
	}

	// order + pagination
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortColumn, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Amount,
			&e.CategoryID,
			&e.UserID,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	// cek error iterasi rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// PENTING: selalu return sesuatu sesuai signature
	return out, nil
}

// ListByUser: shortcut tanpa filter
func (r *ExpenseRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Expense, error) {
	return r.List(ctx, userID, nil, nil, nil, 10, 0, "date", "desc")
}

// Create: tambah expense baru
func (r *ExpenseRepo) Create(ctx context.Context, e *models.Expense) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO expenses (id, title, description, amount, category_id, user_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
		e.ID, e.Title, e.Description, e.Amount, e.CategoryID, e.UserID)
	return err
}

// Update: ubah expense milik user
func (r *ExpenseRepo) Update(ctx context.Context, userID, id uuid.UUID, title, description string, amount float64, categoryID uuid.UUID) (bool, error) {
	ct, err := r.db.Exec(ctx,
		`UPDATE expenses
		 SET title=$1, description=$2, amount=$3, category_id=$4, updated_at=NOW()
		 WHERE id=$5 AND user_id=$6`,
		title, description, amount, categoryID, id, userID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// Delete: hapus expense milik user
func (r *ExpenseRepo) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	ct, err := r.db.Exec(ctx,
		`UPDATE expenses
		 SET deleted_at = NOW()
		 WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`,
		id, userID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
