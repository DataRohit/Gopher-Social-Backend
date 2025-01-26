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

// ErrCommentNotFound is returned when a comment is not found.
var ErrCommentNotFound = errors.New("comment not found")

type CommentStore struct {
	dbPool *pgxpool.Pool
}

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

	return cs.GetCommentByID(ctx, comment.ID)
}

// GetCommentByID retrieves a comment from the database by its ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - commentID (uuid.UUID): ID of the comment to retrieve.
//
// Returns:
//   - *models.Comment: The retrieved comment if found.
//   - error: ErrCommentNotFound if comment not found or other errors during database query.
func (cs *CommentStore) GetCommentByID(ctx context.Context, commentID uuid.UUID) (*models.Comment, error) {
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
			(SELECT COUNT(*) FROM post_likes pd WHERE pd.post_id = p.id AND pd.liked = FALSE) as dislikes_count
		FROM comments c
		INNER JOIN users u ON c.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		INNER JOIN posts p ON c.post_id = p.id
		WHERE c.id = $1
	`, commentID).Scan(
		&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
		&comment.Author.ID, &comment.Author.Username, &comment.Author.Email, &comment.Author.Banned, &comment.Author.IsActive, &comment.Author.CreatedAt, &comment.Author.UpdatedAt,
		&comment.Author.Role.Level, &comment.Author.Role.Description,
		&comment.Author.Followers, &comment.Author.Following,
		&comment.Post.ID, &comment.Post.AuthorID, &comment.Post.Title, &comment.Post.SubTitle, &comment.Post.Description, &comment.Post.Content, &comment.Post.CreatedAt, &comment.Post.UpdatedAt,
		&comment.Post.Likes, &comment.Post.Dislikes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}

	return &comment, nil
}
