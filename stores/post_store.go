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

type PostStore struct {
	dbPool    *pgxpool.Pool
	authStore *AuthStore // AuthStore to fetch user details
}

// NewPostStore creates a new PostStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//   - authStore (*AuthStore): AuthStore instance.
//
// Returns:
//   - *PostStore: PostStore instance.
func NewPostStore(dbPool *pgxpool.Pool, authStore *AuthStore) *PostStore {
	return &PostStore{
		dbPool:    dbPool,
		authStore: authStore,
	}
}

// ErrPostNotFound is returned when a post is not found.
var ErrPostNotFound = errors.New("post not found")

// CreatePost creates a new post in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - post (*models.Post): Post object to be created.
//
// Returns:
//   - *models.Post: The created post if successful.
//   - error: An error if post creation fails.
func (ps *PostStore) CreatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	var createdPost models.Post
	post.ID = uuid.New()

	err := ps.dbPool.QueryRow(ctx, `
		INSERT INTO posts (
			id,
			author_id,
			title,
			sub_title,
			description,
			content,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, author_id, title, sub_title, description, content, created_at, updated_at
	`,
		post.ID, post.AuthorID, post.Title, post.SubTitle, post.Description, post.Content,
	).Scan(
		&createdPost.ID, &createdPost.AuthorID, &createdPost.Title, &createdPost.SubTitle, &createdPost.Description, &createdPost.Content, &createdPost.CreatedAt, &createdPost.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return &createdPost, nil
}

// GetPostByID retrieves a post from the database by its ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post to retrieve.
//
// Returns:
//   - *models.Post: The retrieved post if found.
//   - error: ErrPostNotFound if post not found or other errors during database query.
func (ps *PostStore) GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := ps.dbPool.QueryRow(ctx, `
		SELECT
			id, author_id, title, sub_title, description, content, created_at, updated_at
		FROM posts
		WHERE id = $1
	`, postID).Scan(
		&post.ID, &post.AuthorID, &post.Title, &post.SubTitle, &post.Description, &post.Content, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	return &post, nil
}

// UpdatePost updates an existing post in the database.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - post (*models.Post): Post object with updated information. Post ID must be set.
//
// Returns:
//   - *models.Post: The updated post if successful.
//   - error: ErrPostNotFound if post not found or other errors during database query.
func (ps *PostStore) UpdatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	var updatedPost models.Post
	err := ps.dbPool.QueryRow(ctx, `
		UPDATE posts
		SET
			title = $2,
			sub_title = $3,
			description = $4,
			content = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, author_id, title, sub_title, description, content, created_at, updated_at
	`,
		post.ID, post.Title, post.SubTitle, post.Description, post.Content,
	).Scan(
		&updatedPost.ID, &updatedPost.AuthorID, &updatedPost.Title, &updatedPost.SubTitle, &updatedPost.Description, &updatedPost.Content, &updatedPost.CreatedAt, &updatedPost.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return &updatedPost, nil
}
