package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rifqi535/expense-tracker-api/internal/models"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users(id, name, email, password_hash) VALUES ($1,$2,$3,$4)`,
		u.ID, u.Name, u.Email, u.PasswordHash,
	)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email=$1`, email)
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id=$1`, id)
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
