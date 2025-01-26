package stores

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datarohit/gopher-social-backend/models"
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

// ListTimedOutUsers retrieves a list of users who are currently timed out.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.User: A slice of User pointers, or nil if no users are timed out.
//   - error: An error if the database query fails.
func (as *ActionStore) ListTimedOutUsers(ctx context.Context, pageNumber int, pageSize int) ([]*models.User, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := as.dbPool.Query(ctx, `
		SELECT
			u.id, u.username, u.email, u.timeout_until, u.banned, u.is_active, u.created_at, u.updated_at,
			r.id as role_id, r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM users u
		INNER JOIN roles r ON u.role_id = r.id
		WHERE u.timeout_until > NOW()
		ORDER BY u.timeout_until ASC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list timed out users: %w", err)
	}
	defer rows.Close()

	var timedOutUsers []*models.User
	for rows.Next() {
		timedOutUser := &models.User{Role: &models.Role{}}
		var timeoutUntil time.Time

		err := rows.Scan(
			&timedOutUser.ID, &timedOutUser.Username, &timedOutUser.Email, &timeoutUntil, &timedOutUser.Banned, &timedOutUser.IsActive, &timedOutUser.CreatedAt, &timedOutUser.UpdatedAt,
			&timedOutUser.Role.ID, &timedOutUser.Role.Level, &timedOutUser.Role.Description,
			&timedOutUser.Followers, &timedOutUser.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timed out user row: %w", err)
		}
		timedOutUser.TimeoutUntil = &timeoutUntil
		timedOutUsers = append(timedOutUsers, timedOutUser)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during timed out users rows iteration: %w", err)
	}

	return timedOutUsers, nil
}
