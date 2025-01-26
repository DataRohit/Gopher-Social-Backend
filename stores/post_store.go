package stores

import (
	"context"
	"fmt"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostStore struct {
	dbPool *pgxpool.Pool
}

// NewPostStore creates a new PostStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *PostStore: PostStore instance.
func NewPostStore(dbPool *pgxpool.Pool) *PostStore {
	return &PostStore{
		dbPool: dbPool,
	}
}

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
