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
			p.id, p.author_id, p.title, p.sub_title, p.description, p.content, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.liked = TRUE) as likes_count,
			(SELECT COUNT(*) FROM post_likes pd WHERE pd.post_id = p.id AND pd.liked = FALSE) as dislikes_count
		FROM posts p
		WHERE id = $1
	`, postID).Scan(
		&post.ID, &post.AuthorID, &post.Title, &post.SubTitle, &post.Description, &post.Content, &post.CreatedAt, &post.UpdatedAt,
		&post.Likes, &post.Dislikes,
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

// DeletePost deletes an existing post from the database by its ID.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post to be deleted.
//
// Returns:
//   - error: An error if deleting the post fails or if the post is not found.
func (ps *PostStore) DeletePost(ctx context.Context, postID uuid.UUID) error {
	commandTag, err := ps.dbPool.Exec(ctx, `
		DELETE FROM posts
		WHERE id = $1
	`, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPostNotFound
	}

	return nil
}

// ListPostsByAuthorID retrieves all posts from the database for a given author ID with pagination.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - authorID (uuid.UUID): ID of the author whose posts are to be retrieved.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Page size for pagination.
//
// Returns:
//   - []*models.Post: A slice of Post pointers, or nil if no posts are found.
//   - error: An error if the database query fails.
func (ps *PostStore) ListPostsByAuthorID(ctx context.Context, authorID uuid.UUID, pageNumber int, pageSize int) ([]*models.Post, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := ps.dbPool.Query(ctx, `
		SELECT
			p.id, p.author_id, p.title, p.sub_title, p.description, p.content, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.liked = TRUE) as likes_count,
			(SELECT COUNT(*) FROM post_likes pd WHERE pd.post_id = p.id AND pd.liked = FALSE) as dislikes_count
		FROM posts p
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts by author id: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID, &post.AuthorID, &post.Title, &post.SubTitle, &post.Description, &post.Content, &post.CreatedAt, &post.UpdatedAt,
			&post.Likes, &post.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during posts rows iteration: %w", err)
	}

	return posts, nil
}
