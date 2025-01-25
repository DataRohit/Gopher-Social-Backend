package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore struct {
	dbPool *pgxpool.Pool
}

// NewAuthStore creates a new AuthStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *AuthStore: AuthStore instance.
func NewAuthStore(dbPool *pgxpool.Pool) *AuthStore {
	return &AuthStore{
		dbPool: dbPool,
	}
}

// ErrUserAlreadyExists is returned when a user with the same username or email already exists.
var ErrUserAlreadyExists = errors.New("user already exists")

// CreateUser creates a new user in the database.
// It checks if a user with the same username or email already exists before creating a new user.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - user (*models.User): User object to be created.
//
// Returns:
//   - *models.User: The created user if successful.
//   - error: An error if user creation fails or user already exists.
func (as *AuthStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	var existingUser models.User
	err := as.dbPool.QueryRow(ctx, `SELECT id, username, email, password_hash, timeout_until, banned, created_at, updated_at FROM users WHERE username = $1 OR email = $2`, user.Username, user.Email).Scan(
		&existingUser.ID, &existingUser.Username, &existingUser.Email, &existingUser.PasswordHash, &existingUser.TimeoutUntil, &existingUser.Banned, &existingUser.CreatedAt, &existingUser.UpdatedAt,
	)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}

	var createdUser models.User
	err = as.dbPool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password_hash, timeout_until, banned, created_at, updated_at
		`, user.Username, user.Email, user.PasswordHash).Scan(
		&createdUser.ID, &createdUser.Username, &createdUser.Email, &createdUser.PasswordHash, &createdUser.TimeoutUntil, &createdUser.Banned, &createdUser.CreatedAt, &createdUser.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	createdUser.CreatedAt = helpers.ConvertToAsiaMumbaiTime(createdUser.CreatedAt)
	createdUser.UpdatedAt = helpers.ConvertToAsiaMumbaiTime(createdUser.UpdatedAt)

	return &createdUser, nil
}
