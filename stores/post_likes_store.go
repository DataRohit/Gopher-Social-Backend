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

// PostLikeStore manages post like data operations.
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

// ErrPostDislikeAlreadyExists is returned when a user has already disliked a post.
var ErrPostDislikeAlreadyExists = errors.New("post dislike already exists")

// ErrPostLikeNotFound is returned when a post like is not found.
var ErrPostLikeNotFound = errors.New("post like not found")

// ErrPostDislikeNotFound is returned when a post dislike is not found.
var ErrPostDislikeNotFound = errors.New("post dislike not found")

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
		if existingLike.Liked {
			return nil, ErrPostLikeAlreadyExists
		}

		updatedLike := models.PostLike{}
		err = pls.dbPool.QueryRow(ctx, `
			UPDATE post_likes
			SET liked = TRUE
			WHERE user_id = $1 AND post_id = $2
			RETURNING user_id, post_id, liked, created_at
		`, userID, postID).Scan(
			&updatedLike.UserID, &updatedLike.PostID, &updatedLike.Liked, &updatedLike.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update post like from dislike to like: %w", err)
		}
		return &updatedLike, nil
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

// DislikePost creates a new post dislike record in the database.
// It records that a user has disliked a specific post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who disliked the post.
//   - postID (uuid.UUID): ID of the post that was disliked.
//
// Returns:
//   - *models.PostLike: The created PostLike object if successful.
//   - error: An error if creating the dislike record fails or if the dislike already exists.
func (pls *PostLikeStore) DislikePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*models.PostLike, error) {
	var existingDislike models.PostLike
	err := pls.dbPool.QueryRow(ctx, `SELECT user_id, post_id, liked, created_at FROM post_likes WHERE user_id = $1 AND post_id = $2`, userID, postID).Scan(
		&existingDislike.UserID, &existingDislike.PostID, &existingDislike.Liked, &existingDislike.CreatedAt,
	)
	if err == nil {
		if !existingDislike.Liked {
			return nil, ErrPostDislikeAlreadyExists
		}

		updatedDislike := models.PostLike{}
		err = pls.dbPool.QueryRow(ctx, `
			UPDATE post_likes
			SET liked = FALSE
			WHERE user_id = $1 AND post_id = $2
			RETURNING user_id, post_id, liked, created_at
		`, userID, postID).Scan(
			&updatedDislike.UserID, &updatedDislike.PostID, &updatedDislike.Liked, &updatedDislike.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update post like from like to dislike: %w", err)
		}
		return &updatedDislike, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing post dislike: %w", err)
	}

	var createdDislike models.PostLike
	err = pls.dbPool.QueryRow(ctx, `
		INSERT INTO post_likes (user_id, post_id, liked)
		VALUES ($1, $2, FALSE)
		RETURNING user_id, post_id, liked, created_at
	`, userID, postID).Scan(
		&createdDislike.UserID, &createdDislike.PostID, &createdDislike.Liked, &createdDislike.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dislike post: %w", err)
	}

	return &createdDislike, nil
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
		WHERE user_id = $1 AND post_id = $2 AND liked = TRUE
	`, userID, postID)
	if err != nil {
		return fmt.Errorf("failed to unlike post: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPostLikeNotFound
	}

	return nil
}

// UndislikePost removes a post dislike record from the database.
// It signifies that a user has undisliked a specific post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who is undisliking the post.
//   - postID (uuid.UUID): ID of the post to be undisliked.
//
// Returns:
//   - error: An error if removing the dislike record fails or if the dislike is not found.
func (pls *PostLikeStore) UndislikePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	commandTag, err := pls.dbPool.Exec(ctx, `
		DELETE FROM post_likes
		WHERE user_id = $1 AND post_id = $2 AND liked = FALSE
	`, userID, postID)
	if err != nil {
		return fmt.Errorf("failed to undislike post: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPostDislikeNotFound
	}

	return nil
}

// GetPostLikeByUserAndPost retrieves a post like record by user ID and post ID.
// It checks if a specific user has liked or disliked a specific post.
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
