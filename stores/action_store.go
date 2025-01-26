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

// ErrAdminCannotDeactivateAdmin is returned when an admin tries to deactivate another admin.
var ErrAdminCannotDeactivateAdmin = errors.New("admin cannot deactivate another admin")

// ErrModeratorCannotDeactivateModeratorOrAdmin is returned when a moderator tries to deactivate a moderator or admin.
var ErrModeratorCannotDeactivateModeratorOrAdmin = errors.New("moderator can only deactivate normal users")

// ErrAdminCannotActivateAdmin is returned when an admin tries to activate another admin.
var ErrAdminCannotActivateAdmin = errors.New("admin cannot activate another admin")

// ErrModeratorCannotActivateModeratorOrAdmin is returned when a moderator tries to activate a moderator or admin.
var ErrModeratorCannotActivateModeratorOrAdmin = errors.New("moderator can only activate normal users")

// ErrAdminCannotBanAdmin is returned when an admin tries to ban another admin.
var ErrAdminCannotBanAdmin = errors.New("admin cannot ban another admin")

// ErrAdminCannotUnbanAdmin is returned when an admin tries to unban another admin.
var ErrAdminCannotUnbanAdmin = errors.New("admin cannot unban another admin")


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

// DeactivateUser deactivates a user by setting their is_active status to false.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to deactivate.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) DeactivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	commandTag, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET is_active = FALSE
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ActivateUser activates a user by setting their is_active status to true.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to activate.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) ActivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	commandTag, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET is_active = TRUE
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// BanUser bans a user, deactivates them, deletes their posts, and sets the banned status to true.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to ban.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) BanUser(ctx context.Context, targetUserID uuid.UUID) error {
	tx, err := as.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Deactivate User and set banned to true
	_, err = tx.Exec(ctx, `
		UPDATE users
		SET is_active = FALSE, banned = TRUE
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Delete User's Posts
	_, err = tx.Exec(ctx, `
		DELETE FROM posts
		WHERE author_id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to delete user's posts: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UnbanUser unbans a user by setting their banned status to false.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - targetUserID (uuid.UUID): ID of the user to unban.
//
// Returns:
//   - error: An error if the operation fails.
func (as *ActionStore) UnbanUser(ctx context.Context, targetUserID uuid.UUID) error {
	commandTag, err := as.dbPool.Exec(ctx, `
		UPDATE users
		SET banned = FALSE
		WHERE id = $1
	`, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
