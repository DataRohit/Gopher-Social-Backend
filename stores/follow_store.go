package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowStore struct {
	dbPool *pgxpool.Pool
}

// NewFollowStore creates a new FollowStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *FollowStore: FollowStore instance.
func NewFollowStore(dbPool *pgxpool.Pool) *FollowStore {
	return &FollowStore{
		dbPool: dbPool,
	}
}

// ErrAlreadyFollowing is returned when a user is already following another user.
var ErrAlreadyFollowing = errors.New("already following user")

// ErrNotFollowing is returned when a user is not following another user.
var ErrNotFollowing = errors.New("not following user")

// FollowUser creates a new follow relationship in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - followerID (uuid.UUID): ID of the follower user.
//   - followeeID (uuid.UUID): ID of the followee user.
//
// Returns:
//   - error: An error if creating the follow relationship fails or if already following.
func (fs *FollowStore) FollowUser(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error {
	var existingFollow models.Follow
	err := fs.dbPool.QueryRow(ctx, `SELECT follower_id, followee_id, created_at FROM follows WHERE follower_id = $1 AND followee_id = $2`, followerID, followeeID).Scan(
		&existingFollow.FollowerID, &existingFollow.FolloweeID, &existingFollow.CreatedAt,
	)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		return ErrAlreadyFollowing
	}

	_, err = fs.dbPool.Exec(ctx, `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
	`, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}
	return nil
}

// UnfollowUser removes a follow relationship from the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - followerID (uuid.UUID): ID of the follower user.
//   - followeeID (uuid.UUID): ID of the followee user.
//
// Returns:
//   - error: An error if removing the follow relationship fails or if not following.
func (fs *FollowStore) UnfollowUser(ctx context.Context, followerID uuid.UUID, followeeID uuid.UUID) error {
	var existingFollow models.Follow
	err := fs.dbPool.QueryRow(ctx, `SELECT follower_id, followee_id, created_at FROM follows WHERE follower_id = $1 AND followee_id = $2`, followerID, followeeID).Scan(
		&existingFollow.FollowerID, &existingFollow.FolloweeID, &existingFollow.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFollowing
	} else if err != nil {
		return fmt.Errorf("failed to check existing follow: %w", err)
	}

	_, err = fs.dbPool.Exec(ctx, `
		DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2
	`, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}
	return nil
}

// GetFollowersByUserID retrieves all followers of a user, excluding banned users and includes follower/following counts.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - followeeID (uuid.UUID): ID of the user to get followers for.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.User: List of users following the user (followee) with follower and following counts.
//   - error: An error if fetching followers fails.
func (fs *FollowStore) GetFollowersByUserID(ctx context.Context, followeeID uuid.UUID, pageNumber int, pageSize int) ([]*models.User, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := fs.dbPool.Query(ctx, `
		SELECT
			u.id, u.username, u.email, u.role_id, u.timeout_until, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM follows f
		INNER JOIN users u ON f.follower_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE f.followee_id = $1 AND u.banned = FALSE AND u.is_active = TRUE
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`, followeeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var followers []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.RoleID, &user.TimeoutUntil, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
			&user.Role.Level, &user.Role.Description,
			&user.Followers, &user.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower row: %w", err)
		}
		followers = append(followers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during followers rows iteration: %w", err)
	}

	return followers, nil
}

// GetFollowingByUserID retrieves all users being followed by a user, excluding banned users and includes follower/following counts.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - followerID (uuid.UUID): ID of the user to get following users for.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.User: List of users being followed by the user (follower) with follower and following counts.
//   - error: An error if fetching following users fails.
func (fs *FollowStore) GetFollowingByUserID(ctx context.Context, followerID uuid.UUID, pageNumber int, pageSize int) ([]*models.User, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := fs.dbPool.Query(ctx, `
		SELECT
			u.id, u.username, u.email, u.role_id, u.timeout_until, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM follows f
		INNER JOIN users u ON f.followee_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE f.follower_id = $1 AND u.banned = FALSE AND u.is_active = TRUE
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`, followerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following users: %w", err)
	}
	defer rows.Close()

	var following []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.RoleID, &user.TimeoutUntil, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
			&user.Role.Level, &user.Role.Description,
			&user.Followers, &user.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan following user row: %w", err)
		}
		following = append(following, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during following users rows iteration: %w", err)
	}

	return following, nil
}
