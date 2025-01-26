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

type CommentLikeStore struct {
	dbPool *pgxpool.Pool
}

// NewCommentLikeStore creates a new CommentLikeStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *CommentLikeStore: CommentLikeStore instance.
func NewCommentLikeStore(dbPool *pgxpool.Pool) *CommentLikeStore {
	return &CommentLikeStore{
		dbPool: dbPool,
	}
}

// ErrCommentLikeAlreadyExists is returned when a user has already liked a comment.
var ErrCommentLikeAlreadyExists = errors.New("comment like already exists")

// ErrCommentDislikeAlreadyExists is returned when a user has already disliked a comment.
var ErrCommentDislikeAlreadyExists = errors.New("comment dislike already exists")

// ErrCommentLikeNotFound is returned when a comment like is not found.
var ErrCommentLikeNotFound = errors.New("comment like not found")

// ErrCommentDislikeNotFound is returned when a comment dislike is not found.
var ErrCommentDislikeNotFound = errors.New("comment dislike not found")

// LikeComment creates a new comment like record in the database.
// It records that a user has liked a specific comment.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who liked the comment.
//   - commentID (uuid.UUID): ID of the comment that was liked.
//
// Returns:
//   - *models.CommentLike: The created CommentLike object if successful.
//   - error: An error if creating the like record fails or if the like already exists.
func (cls *CommentLikeStore) LikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (*models.CommentLike, error) {
	var existingLike models.CommentLike
	err := cls.dbPool.QueryRow(ctx, `SELECT user_id, comment_id, liked, created_at FROM comment_likes WHERE user_id = $1 AND comment_id = $2`, userID, commentID).Scan(
		&existingLike.UserID, &existingLike.CommentID, &existingLike.Liked, &existingLike.CreatedAt,
	)
	if err == nil {
		if existingLike.Liked {
			return nil, ErrCommentLikeAlreadyExists
		}

		updatedLike := models.CommentLike{}
		err = cls.dbPool.QueryRow(ctx, `
			UPDATE comment_likes
			SET liked = TRUE
			WHERE user_id = $1 AND comment_id = $2
			RETURNING user_id, comment_id, liked, created_at
		`, userID, commentID).Scan(
			&updatedLike.UserID, &updatedLike.CommentID, &updatedLike.Liked, &updatedLike.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update comment like from dislike to like: %w", err)
		}
		return &updatedLike, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing comment like: %w", err)
	}

	var createdLike models.CommentLike
	err = cls.dbPool.QueryRow(ctx, `
		INSERT INTO comment_likes (user_id, comment_id, liked)
		VALUES ($1, $2, TRUE)
		RETURNING user_id, comment_id, liked, created_at
	`, userID, commentID).Scan(
		&createdLike.UserID, &createdLike.CommentID, &createdLike.Liked, &createdLike.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to like comment: %w", err)
	}

	return &createdLike, nil
}

// UnlikeComment removes a comment like record from the database.
// It signifies that a user has unliked a specific comment.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who is unliking the comment.
//   - commentID (uuid.UUID): ID of the comment to be unliked.
//
// Returns:
//   - error: An error if removing the like record fails or if the like is not found.
func (cls *CommentLikeStore) UnlikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error {
	commandTag, err := cls.dbPool.Exec(ctx, `
		DELETE FROM comment_likes
		WHERE user_id = $1 AND comment_id = $2 AND liked = TRUE
	`, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to unlike comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCommentLikeNotFound
	}

	return nil
}
