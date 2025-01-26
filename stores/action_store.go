package stores

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActionStore struct {
	dbPool *pgxpool.Pool
}

// NewActionStore creates a new ActionStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *ActionStore: ActionStore instance.
func NewActionStore(dbPool *pgxpool.Pool) *ActionStore {
	return &ActionStore{
		dbPool: dbPool,
	}
}

// ErrAdminCannotTimeoutAdmin is returned when an admin tries to timeout another admin.
var ErrAdminCannotTimeoutAdmin = errors.New("admin cannot timeout another admin")

// ErrModeratorCannotTimeoutModeratorOrAdmin is returned when a moderator tries to timeout a moderator or admin.
var ErrModeratorCannotTimeoutModeratorOrAdmin = errors.New("moderator can only timeout normal users")

// TimeoutUser applies a timeout to a user until the specified time.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to timeout.
//   - timeoutDuration (time.Duration): Duration of the timeout.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) TimeoutUser(ctx context.Context, targetUserID uuid.UUID, timeoutDuration time.Duration) error {
	expiryTime := time.Now().Add(timeoutDuration)

	commandTag, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET timeout_until = $2
		WHERE id = $1
	`, targetUserID, expiryTime)
	if err != nil {
		return fmt.Errorf("failed to timeout user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// RemoveTimeoutUser removes the timeout from a user.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to remove timeout from.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) RemoveTimeoutUser(ctx context.Context, targetUserID uuid.UUID) error {
	commandTag, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET timeout_until = NULL
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to remove timeout from user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
