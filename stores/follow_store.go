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
