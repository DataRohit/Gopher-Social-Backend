package stores

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/google/uuid"
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

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidOrExpiredToken is returned when a password reset token is invalid or expired.
var ErrInvalidOrExpiredToken = errors.New("invalid or expired reset token")

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

	return &createdUser, nil
}

// GetUserByUsernameOrEmail retrieves a user from the database by username or email.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - identifier (string): Username or email to identify the user.
//
// Returns:
//   - *models.User: The retrieved user if found.
//   - error: ErrUserNotFound if user not found or other errors during database query.
func (as *AuthStore) GetUserByUsernameOrEmail(ctx context.Context, identifier string) (*models.User, error) {
	var user models.User
	err := as.dbPool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, timeout_until, banned, created_at, updated_at, password_reset_token, reset_token_expiry
		FROM users
		WHERE username = $1 OR email = $1
	`, identifier).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.TimeoutUntil, &user.Banned, &user.CreatedAt, &user.UpdatedAt, &user.PasswordResetToken, &user.ResetTokenExpiry,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username or email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user from the database by ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - id (uuid.UUID): User ID to identify the user.
//
// Returns:
//   - *models.User: The retrieved user if found.
//   - error: ErrUserNotFound if user not found or other errors during database query.
func (as *AuthStore) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := as.dbPool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, timeout_until, banned, created_at, updated_at, password_reset_token, reset_token_expiry
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.TimeoutUntil, &user.Banned, &user.CreatedAt, &user.UpdatedAt, &user.PasswordResetToken, &user.ResetTokenExpiry,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// CreatePasswordResetToken stores a password reset token and its expiry time for a user.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): User ID for whom to create the reset token.
//   - token (string): The password reset token.
//   - expiryTime (time.Time): The expiry time of the token.
//
// Returns:
//   - error: An error if storing the token fails.
func (as *AuthStore) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiryTime time.Time) error {
	_, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET password_reset_token = $2, reset_token_expiry = $3
		WHERE id = $1
	`, userID, token, expiryTime)
	if err != nil {
		return fmt.Errorf("failed to store password reset token: %w", err)
	}
	return nil
}

// ValidatePasswordResetToken retrieves and validates a password reset token.
// It checks if the token exists, is not expired, and returns the associated user ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - tokenString (string): The password reset token string.
//   - currentTime (time.Time): The current time to check for expiry.
//
// Returns:
//   - uuid.UUID: The user ID associated with the token if valid.
//   - error: ErrInvalidOrExpiredToken if the token is invalid or expired, or other errors.
func (as *AuthStore) ValidatePasswordResetToken(ctx context.Context, tokenString string, currentTime time.Time) (uuid.UUID, error) {
	var userID uuid.UUID
	var expiryTime time.Time
	var storedToken string

	err := as.dbPool.QueryRow(ctx, `
		SELECT id, reset_token_expiry, password_reset_token
		FROM users
		WHERE password_reset_token = $1
	`, tokenString).Scan(
		&userID, &expiryTime, &storedToken,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidOrExpiredToken
		}
		return uuid.Nil, fmt.Errorf("failed to retrieve password reset token: %w", err)
	}

	if currentTime.After(expiryTime) {
		return uuid.Nil, ErrInvalidOrExpiredToken
	}

	if storedToken != tokenString {
		return uuid.Nil, ErrInvalidOrExpiredToken
	}

	return userID, nil
}

// InvalidatePasswordResetToken invalidates a password reset token by setting it to NULL in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - tokenString (string): The password reset token string to invalidate.
//
// Returns:
//   - error: An error if invalidating the token fails.
func (as *AuthStore) InvalidatePasswordResetToken(ctx context.Context, tokenString string) error {
	_, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET password_reset_token = NULL, reset_token_expiry = NULL
		WHERE password_reset_token = $1
	`, tokenString)
	if err != nil {
		return fmt.Errorf("failed to invalidate password reset token: %w", err)
	}
	return nil
}

// UpdateUserPassword updates a user's password in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user whose password needs to be updated.
//   - hashedPassword (string): The new hashed password.
//
// Returns:
//   - error: An error if updating the password fails.
func (as *AuthStore) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	_, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1
	`, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}
