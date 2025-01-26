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

type PostLikeStore struct {
	dbPool *pgxpool.Pool
}

// NewPostLikeStore creates a new PostLikeStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *PostLikeStore: PostLikeStore instance.
func NewPostLikeStore(dbPool *pgxpool.Pool) *PostLikeStore {
	return &PostLikeStore{
		dbPool: dbPool,
	}
}

// ErrPostLikeAlreadyExists is returned when a user has already liked a post.
var ErrPostLikeAlreadyExists = errors.New("post like already exists")

// ErrPostLikeNotFound is returned when a post like is not found.
var ErrPostLikeNotFound = errors.New("post like not found")

// LikePost creates a new post like record in the database.
// It records that a user has liked a specific post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who liked the post.
//   - postID (uuid.UUID): ID of the post that was liked.
//
// Returns:
//   - *models.PostLike: The created PostLike object if successful.
//   - error: An error if creating the like record fails or if the like already exists.
func (pls *PostLikeStore) LikePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*models.PostLike, error) {
	var existingLike models.PostLike
	err := pls.dbPool.QueryRow(ctx, `SELECT user_id, post_id, liked, created_at FROM post_likes WHERE user_id = $1 AND post_id = $2`, userID, postID).Scan(
		&existingLike.UserID, &existingLike.PostID, &existingLike.Liked, &existingLike.CreatedAt,
	)
	if err == nil {
		return nil, ErrPostLikeAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing post like: %w", err)
	}

	var createdLike models.PostLike
	err = pls.dbPool.QueryRow(ctx, `
		INSERT INTO post_likes (user_id, post_id, liked)
		VALUES ($1, $2, TRUE)
		RETURNING user_id, post_id, liked, created_at
	`, userID, postID).Scan(
		&createdLike.UserID, &createdLike.PostID, &createdLike.Liked, &createdLike.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to like post: %w", err)
	}

	return &createdLike, nil
}

// UnlikePost removes a post like record from the database.
// It signifies that a user has unliked a specific post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who is unliking the post.
//   - postID (uuid.UUID): ID of the post to be unliked.
//
// Returns:
//   - error: An error if removing the like record fails or if the like is not found.
func (pls *PostLikeStore) UnlikePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	commandTag, err := pls.dbPool.Exec(ctx, `
		DELETE FROM post_likes
		WHERE user_id = $1 AND post_id = $2
	`, userID, postID)
	if err != nil {
		return fmt.Errorf("failed to unlike post: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPostLikeNotFound
	}

	return nil
}

// GetPostLikeByUserAndPost retrieves a post like record by user ID and post ID.
// It checks if a specific user has liked a specific post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user.
//   - postID (uuid.UUID): ID of the post.
//
// Returns:
//   - *models.PostLike: The PostLike object if found.
//   - error: ErrPostLikeNotFound if the like record is not found or other errors during database query.
func (pls *PostLikeStore) GetPostLikeByUserAndPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*models.PostLike, error) {
	var postLike models.PostLike
	err := pls.dbPool.QueryRow(ctx, `
		SELECT user_id, post_id, liked, created_at
		FROM post_likes
		WHERE user_id = $1 AND post_id = $2
	`, userID, postID).Scan(
		&postLike.UserID, &postLike.PostID, &postLike.Liked, &postLike.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostLikeNotFound
		}
		return nil, fmt.Errorf("failed to get post like by user and post: %w", err)
	}

	return &postLike, nil
}
