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

type CommentStore struct {
	dbPool *pgxpool.Pool
}

// ErrCommentNotFound is returned when a comment is not found.
var ErrCommentNotFound = errors.New("comment not found")

// NewCommentStore creates a new CommentStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *CommentStore: CommentStore instance.
func NewCommentStore(dbPool *pgxpool.Pool) *CommentStore {
	return &CommentStore{
		dbPool: dbPool,
	}
}

// CreateComment creates a new comment in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - comment (*models.Comment): Comment object to be created.
//
// Returns:
//   - *models.Comment: The created comment if successful.
//   - error: An error if comment creation fails.
func (cs *CommentStore) CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	comment.ID = uuid.New()
	_, err := cs.dbPool.Exec(ctx, `
		INSERT INTO comments (
			id,
			author_id,
			post_id,
			content
		) VALUES ($1, $2, $3, $4)
	`, comment.ID, comment.AuthorID, comment.PostID, comment.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return cs.GetCommentByID(ctx, comment.ID, comment.PostID)
}

// GetCommentByID retrieves a comment from the database by its ID and Post ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - commentID (uuid.UUID): ID of the comment to retrieve.
//   - postID (uuid.UUID): ID of the post to which the comment belongs.
//
// Returns:
//   - *models.Comment: The retrieved comment if found.
//   - error: ErrCommentNotFound if comment not found or other errors during database query.
func (cs *CommentStore) GetCommentByID(ctx context.Context, commentID uuid.UUID, postID uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	comment.Author = &models.User{}
	comment.Author.Role = &models.Role{}
	comment.Post = &models.Post{}

	err := cs.dbPool.QueryRow(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			p.id, p.author_id, p.title, p.sub_title, p.description, p.content, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.liked = TRUE) as likes_count,
			(SELECT COUNT(*) FROM post_likes pd WHERE pd.post_id = p.id AND pd.liked = FALSE) as dislikes_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.liked = TRUE) as likes,
			(SELECT COUNT(*) FROM comment_likes cd WHERE cd.comment_id = c.id AND cd.liked = FALSE) as dislikes
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		INNER JOIN posts p ON c.post_id = p.id
		WHERE c.id = $1 AND p.id = $2
	`, commentID, postID).Scan(
		&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
		&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
		&comment.Author.Role.Level, &comment.Author.Role.Description,
		&comment.Author.Followers, &comment.Author.Following,
		&comment.Post.ID, &comment.Post.AuthorID, &comment.Post.Title, &comment.Post.SubTitle, &comment.Post.Description, &comment.Post.Content, &comment.Post.CreatedAt, &comment.Post.UpdatedAt,
		&comment.Post.Likes, &comment.Post.Dislikes,
		&comment.Likes, &comment.Dislikes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}

	return &comment, nil
}

// UpdateComment updates an existing comment in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - comment (*models.Comment): Comment object with updated information. Comment ID must be populated.
//
// Returns:
//   - *models.Comment: The updated comment if successful.
//   - error: ErrCommentNotFound if comment not found or other errors during database query.
func (cs *CommentStore) UpdateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	commandTag, err := cs.dbPool.Exec(ctx, `
		UPDATE comments
		SET
			content = $2,
			updated_at = NOW()
		WHERE id = $1
	`, comment.ID, comment.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return nil, ErrCommentNotFound
	}

	return cs.GetCommentByID(ctx, comment.ID, comment.PostID)
}

// DeleteComment deletes a comment from the database by its ID and Post ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - commentID (uuid.UUID): ID of the comment to delete.
//   - postID (uuid.UUID): ID of the post to which the comment belongs.
//
// Returns:
//   - error: An error if deletion fails, nil if successful.
func (cs *CommentStore) DeleteComment(ctx context.Context, commentID uuid.UUID, postID uuid.UUID) error {
	commandTag, err := cs.dbPool.Exec(ctx, `
		DELETE FROM comments
		WHERE id = $1 AND post_id = $2
	`, commentID, postID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// ListCommentsByAuthorIDForPost retrieves all comments for a given post made by a specific author from the database with pagination.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - authorID (uuid.UUID): ID of the author of comments.
//   - postID (uuid.UUID): ID of the post to which the comments belongs.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of comments per page.
//
// Returns:
//   - []*models.Comment: List of comments if found.
//   - error: An error if retrieval fails.
func (cs *CommentStore) ListCommentsByAuthorIDForPost(ctx context.Context, authorID uuid.UUID, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	var comments []*models.Comment
	offset := (pageNumber - 1) * pageSize

	rows, err := cs.dbPool.Query(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.liked = TRUE) as likes,
			(SELECT COUNT(*) FROM comment_likes cd WHERE cd.comment_id = c.id AND cd.liked = FALSE) as dislikes
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE c.author_id = $1 AND c.post_id = $2
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4
	`, authorID, postID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by author for post: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		comment := &models.Comment{}
		comment.Author = &models.User{}
		comment.Author.Role = &models.Role{}
		if err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
			&comment.Author.Role.Level, &comment.Author.Role.Description,
			&comment.Author.Followers, &comment.Author.Following,
			&comment.Likes, &comment.Dislikes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during comments rows iteration: %w", err)
	}

	return comments, nil
}

// ListCommentsByUserIdentifierForPost retrieves all comments for a given post made by a user identifier (username or email or userID) from the database with pagination.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - identifier (string):  Username or Email or UserID of the author of comments.
//   - postID (uuid.UUID): ID of the post to which the comments belongs.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of comments per page.
//
// Returns:
//   - []*models.Comment: List of comments if found.
//   - error: An error if retrieval fails.
func (cs *CommentStore) ListCommentsByUserIdentifierForPost(ctx context.Context, identifier string, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	authStore := NewAuthStore(cs.dbPool) // Create a new AuthStore instance
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

	return cs.ListCommentsByAuthorIDForPost(ctx, user.ID, postID, pageNumber, pageSize)
}

// ListCommentsByPostID retrieves all comments for a given post from the database with pagination, ordered by creation time.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of comments per page.
//
// Returns:
//   - []*models.Comment: List of comments if found.
//   - error: An error if retrieval fails.
func (cs *CommentStore) ListCommentsByPostID(ctx context.Context, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	return cs.listCommentsByPostIDOrdered(ctx, postID, pageNumber, pageSize, "c.created_at ASC")
}

// ListCommentsByPostIDLatestFirst retrieves all comments for a given post from the database with pagination, ordered by creation time, latest first.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of comments per page.
//
// Returns:
//   - []*models.Comment: List of comments if found.
//   - error: An error if retrieval fails.
func (cs *CommentStore) ListCommentsByPostIDLatestFirst(ctx context.Context, postID uuid.UUID, pageNumber int, pageSize int) ([]*models.Comment, error) {
	return cs.listCommentsByPostIDOrdered(ctx, postID, pageNumber, pageSize, "c.created_at DESC")
}

// listCommentsByPostIDOrdered is a helper function to retrieve comments for a given post from the database with pagination and custom ordering.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of comments per page.
//   - orderBy (string): SQL order by clause.
//
// Returns:
//   - []*models.Comment: List of comments if found.
//   - error: An error if retrieval fails.
func (cs *CommentStore) listCommentsByPostIDOrdered(ctx context.Context, postID uuid.UUID, pageNumber int, pageSize int, orderBy string) ([]*models.Comment, error) {
	var comments []*models.Comment
	offset := (pageNumber - 1) * pageSize

	rows, err := cs.dbPool.Query(ctx, `
		SELECT
			c.id, c.author_id, c.post_id, c.content, c.created_at, c.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.liked = TRUE) as likes,
			(SELECT COUNT(*) FROM comment_likes cd WHERE cd.comment_id = c.id AND cd.liked = FALSE) as dislikes
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		WHERE c.post_id = $1
		ORDER BY `+orderBy+`
		LIMIT $2 OFFSET $3
	`, postID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by post id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		comment := &models.Comment{}
		comment.Author = &models.User{}
		comment.Author.Role = &models.Role{}
		if err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
			&comment.Author.Role.Level, &comment.Author.Role.Description,
			&comment.Author.Followers, &comment.Author.Following,
			&comment.Likes, &comment.Dislikes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during comments rows iteration: %w", err)
	}

	return comments, nil
}
