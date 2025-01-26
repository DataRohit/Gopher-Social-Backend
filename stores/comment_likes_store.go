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

// DislikeComment creates a new comment dislike record in the database.
// It records that a user has disliked a specific comment.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who disliked the comment.
//   - commentID (uuid.UUID): ID of the comment that was disliked.
//
// Returns:
//   - *models.CommentLike: The created CommentLike object if successful.
//   - error: An error if creating the dislike record fails or if the dislike already exists.
func (cls *CommentLikeStore) DislikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (*models.CommentLike, error) {
	var existingDislike models.CommentLike
	err := cls.dbPool.QueryRow(ctx, `SELECT user_id, comment_id, liked, created_at FROM comment_likes WHERE user_id = $1 AND comment_id = $2`, userID, commentID).Scan(
		&existingDislike.UserID, &existingDislike.CommentID, &existingDislike.Liked, &existingDislike.CreatedAt,
	)
	if err == nil {
		if !existingDislike.Liked {
			return nil, ErrCommentDislikeAlreadyExists
		}

		updatedDislike := models.CommentLike{}
		err = cls.dbPool.QueryRow(ctx, `
			UPDATE comment_likes
			SET liked = FALSE
			WHERE user_id = $1 AND comment_id = $2
			RETURNING user_id, comment_id, liked, created_at
		`, userID, commentID).Scan(
			&updatedDislike.UserID, &updatedDislike.CommentID, &updatedDislike.Liked, &updatedDislike.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update comment like from like to dislike: %w", err)
		}
		return &updatedDislike, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing comment dislike: %w", err)
	}

	var createdDislike models.CommentLike
	err = cls.dbPool.QueryRow(ctx, `
		INSERT INTO comment_likes (user_id, comment_id, liked)
		VALUES ($1, $2, FALSE)
		RETURNING user_id, comment_id, liked, created_at
	`, userID, commentID).Scan(
		&createdDislike.UserID, &createdDislike.CommentID, &createdDislike.Liked, &createdDislike.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dislike comment: %w", err)
	}

	return &createdDislike, nil
}

// UndislikeComment removes a comment dislike record from the database.
// It signifies that a user has removed dislike from a specific comment.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user who is removing dislike from the comment.
//   - commentID (uuid.UUID): ID of the comment to be undisliked.
//
// Returns:
//   - error: An error if removing the dislike record fails or if the dislike is not found.
func (cls *CommentLikeStore) UndislikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error {
	commandTag, err := cls.dbPool.Exec(ctx, `
		DELETE FROM comment_likes
		WHERE user_id = $1 AND comment_id = $2 AND liked = FALSE
	`, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to undislike comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCommentDislikeNotFound
	}

	return nil
}

// ListLikedCommentsByUserIDForPost retrieves all comments liked by a user under a specific post from the database with pagination.
// It returns a list of comments with author information.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.Comment: A slice of Comment pointers, or nil if no comments are found for the given user ID and post ID.
//   - error: An error if fetching the comments fails.
func (cls *CommentLikeStore) ListLikedCommentsByUserIDForPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := cls.dbPool.Query(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM comment_likes cl
		INNER JOIN comments c ON cl.comment_id = c.id
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE cl.user_id = $1 AND c.post_id = $2 AND cl.liked = TRUE
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, postID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list liked comments for post: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{Author: &models.User{Role: &models.Role{}}}
		err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
			&comment.Author.Role.Level, &comment.Author.Role.Description,
			&comment.Author.Followers, &comment.Author.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during comments rows iteration: %w", err)
	}

	return comments, nil
}

// ListDislikedCommentsByUserIDForPost retrieves all comments disliked by a user under a specific post from the database with pagination.
// It returns a list of comments with author information.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.Comment: A slice of Comment pointers, or nil if no comments are found for the given user ID and post ID.
//   - error: An error if fetching the comments fails.
func (cls *CommentLikeStore) ListDislikedCommentsByUserIDForPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := cls.dbPool.Query(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM comment_likes cl
		INNER JOIN comments c ON cl.comment_id = c.id
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE cl.user_id = $1 AND c.post_id = $2 AND cl.liked = FALSE
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, postID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list disliked comments for post: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{Author: &models.User{Role: &models.Role{}}}
		err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
			&comment.Author.Role.Level, &comment.Author.Role.Description,
			&comment.Author.Followers, &comment.Author.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during comments rows iteration: %w", err)
	}

	return comments, nil
}

// ListLikedCommentsByUserIdentifierForPost retrieves all comments liked by a user identifier under a specific post from the database with pagination.
// It resolves the user identifier to a user ID and then fetches the liked comments for that user.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - identifier (string): Username or email of the user.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.Comment: A slice of Comment pointers, or nil if no comments are found for the given user identifier and post ID.
//   - error: An error if fetching the comments fails.
func (cls *CommentLikeStore) ListLikedCommentsByUserIdentifierForPost(ctx context.Context, identifier string, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	authStore := NewAuthStore(cls.dbPool)
	user, err := authStore.GetUserByUsernameOrEmail(ctx, identifier)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			userID, uuidErr := uuid.Parse(identifier)
			if uuidErr != nil {
				return nil, ErrUserNotFound
			}
			user, err = authStore.GetUserByID(ctx, userID)
			if err != nil {
				return nil, ErrUserNotFound
			}
		} else {
			return nil, fmt.Errorf("failed to get user by identifier: %w", err)
		}
	}

	return cls.listLikedCommentsByUserStatusForPost(ctx, user.ID, postID, pageNumber, pageSize, true)
}

// listLikedCommentsByUserStatusForPost is a helper function to retrieve comments based on like status (liked or disliked) for a user identified by user id under a post with pagination.
// It is used by ListLikedCommentsByUserIdentifierForPost and ListDislikedCommentsByUserIdentifierForPost to avoid code duplication.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - userID (uuid.UUID): ID of the user.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//   - liked (bool): True to retrieve liked comments, false for disliked comments.
//
// Returns:
//   - []*models.Comment: A slice of Comment pointers, or nil if no comments are found for the given like status and user identifier.
//   - error: ErrUserNotFound if user is not found, or other errors during database query.
func (cls *CommentLikeStore) listLikedCommentsByUserStatusForPost(ctx context.Context, userID uuid.UUID, postID uuid.UUID, pageNumber int, pageSize int, liked bool) ([]*models.Comment, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := cls.dbPool.Query(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM comment_likes cl
		INNER JOIN comments c ON cl.comment_id = c.id
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE cl.user_id = $1 AND c.post_id = $2 AND cl.liked = $3
		ORDER BY c.created_at DESC
		LIMIT $4 OFFSET $5
	`, userID, postID, liked, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list liked comments for post: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{Author: &models.User{Role: &models.Role{}}}
		err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
			&comment.Author.Role.Level, &comment.Author.Role.Description,
			&comment.Author.Followers, &comment.Author.Following,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during comments rows iteration: %w", err)
	}

	return comments, nil
}
