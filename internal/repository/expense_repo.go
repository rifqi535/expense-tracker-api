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

// List: ambil data expenses dengan filter opsional
func (r *ExpenseRepo) List(
	ctx context.Context,
	userID uuid.UUID,
	categoryID *uuid.UUID,
	startDate, endDate *time.Time,
) ([]models.Expense, error) {

	where := []string{"user_id = $1"}
	params := []interface{}{userID}
	paramIndex := 2

	if categoryID != nil {
		where = append(where, fmt.Sprintf("category_id = $%d", paramIndex))
		params = append(params, *categoryID)
		paramIndex++
	}
	if startDate != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", paramIndex))
		params = append(params, *startDate)
		paramIndex++
	}
	if endDate != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", paramIndex))
		params = append(params, *endDate)
		paramIndex++
	}

	query := `
		SELECT id, title, description, amount, category_id, user_id, created_at, updated_at
		FROM expenses
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Amount,
			&e.CategoryID,
			&e.UserID,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	return expenses, rows.Err()
}

// ListByUser: shortcut tanpa filter
func (r *ExpenseRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Expense, error) {
	return r.List(ctx, userID, nil, nil, nil)
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
		`DELETE FROM expenses WHERE id=$1 AND user_id=$2`,
		id, userID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
