package stores

import (
	"context"
	"fmt"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedStore struct {
	dbPool *pgxpool.Pool
}

// NewFeedStore creates a new FeedStore.
//
// Parameters:
//   - dbPool (*pgxpool.Pool): Pgx connection pool.
//
// Returns:
//   - *FeedStore: FeedStore instance.
func NewFeedStore(dbPool *pgxpool.Pool) *FeedStore {
	return &FeedStore{
		dbPool: dbPool,
	}
}

// ListLatestPosts retrieves the latest posts from the database with pagination for the feed.
// It includes author information, follower/following counts, and like/dislike counts for each post.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - pageNumber (int): Page number for pagination.
//   - pageSize (int): Number of posts per page.
//
// Returns:
//   - []*models.Post: A slice of Post pointers containing the latest posts with details.
//   - error: An error if the database query fails.
func (fs *FeedStore) ListLatestPosts(ctx context.Context, pageNumber int, pageSize int) ([]*models.Post, error) {
	offset := (pageNumber - 1) * pageSize
	rows, err := fs.dbPool.Query(ctx, `
		SELECT
			p.id, p.author_id, p.title, p.sub_title, p.description, p.content, p.created_at, p.updated_at,
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.liked = TRUE) as likes_count,
			(SELECT COUNT(*) FROM post_likes pd WHERE pd.post_id = p.id AND pd.liked = FALSE) as dislikes_count,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM posts p
		INNER JOIN users u ON p.author_id = u.id
		INNER JOIN roles r ON u.role_id = r.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list latest posts: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{Author: &models.User{Role: &models.Role{}}}
		err := rows.Scan(
			&post.ID, &post.AuthorID, &post.Title, &post.SubTitle, &post.Description, &post.Content, &post.CreatedAt, &post.UpdatedAt,
			&post.Author.ID, &post.Author.Username, &post.Author.Email, &post.Author.Banned, &post.Author.IsActive, &post.Author.CreatedAt, &post.Author.UpdatedAt,
			&post.Author.Role.Level, &post.Author.Role.Description,
			&post.Likes, &post.Dislikes,
			&post.Author.Followers, &post.Author.Following,
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
