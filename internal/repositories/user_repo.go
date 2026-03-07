package repositories

import (
	"context"
	"errors"
	"fmt"

	"messaging-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u models.User) (models.User, error) {
	q := `
INSERT INTO users (id, username, password_hash)
VALUES ($1, $2, $3)
RETURNING created_at`
	createdAt := u.CreatedAt
	err := r.db.QueryRow(ctx, q, u.ID, u.Username, u.PasswordHash).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrConflict
		}
		return models.User{}, fmt.Errorf("user create: %w", err)
	}
	u.CreatedAt = createdAt
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	q := `SELECT id, username, password_hash, created_at FROM users WHERE id=$1`
	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("user get by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (models.User, error) {
	q := `SELECT id, username, password_hash, created_at FROM users WHERE username=$1`
	var u models.User
	err := r.db.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("user get by username: %w", err)
	}
	return u, nil
}
