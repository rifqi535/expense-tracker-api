package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rifqi535/expense-tracker-api/internal/models"
)

type CategoryRepo struct{ db *pgxpool.Pool }

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, title, user_id, created_at, updated_at FROM categories WHERE user_id=$1 ORDER BY title`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Title, &c.UserID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) Create(ctx context.Context, c *models.Category) error {
	_, err := r.db.Exec(ctx, `INSERT INTO categories(id, title, user_id) VALUES ($1,$2,$3)`, c.ID, c.Title, c.UserID)
	return err
}

func (r *CategoryRepo) Update(ctx context.Context, userID, id uuid.UUID, title string) (bool, error) {
	ct, err := r.db.Exec(ctx, `
        UPDATE categories 
        SET title=$1, updated_at=NOW() 
        WHERE id=$2 AND user_id=$3`,
		title, id, userID,
	)
	if err != nil {
		return false, err
	}

	// Kalau tidak ada baris yang kena update, return false
	return ct.RowsAffected() > 0, nil
}

func (r *CategoryRepo) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	// Cegah hapus bila dipakai? (opsional) Di sini biarkan DB constraint yang bicara.
	ct, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
