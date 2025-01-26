package stores

import (
	"context"
	"fmt"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/google/uuid"
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

// GetPostWithComments retrieves a specific post by postID along with its comments in paginated form for the feed.
// It includes post details, author information, comment details, and comment author information.
//
// Parameters:
//   - ctx (context.Context): Context for the database operation.
//   - postID (uuid.UUID): ID of the post to retrieve.
//   - pageNumber (int): Page number for comments pagination.
//   - pageSize (int): Number of comments per page.
//
// Returns:
//   - *models.FeedPost: A FeedPost object containing the post and its comments.
//   - error: An error if the database query fails or post is not found.
func (fs *FeedStore) GetPostWithComments(ctx context.Context, postID uuid.UUID, pageNumber int, pageSize int) (*models.FeedPost, error) {
	postStore := NewPostStore(fs.dbPool)
	commentStore := NewCommentStore(fs.dbPool)

	retrievedPost, err := postStore.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	comments, err := commentStore.ListCommentsByPostIDLatestFirst(ctx, postID, pageNumber, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments for post: %w", err)
	}

	feedPost := &models.FeedPost{
		Post:     retrievedPost,
		Comments: comments,
	}

	postAuthor, err := fs.getAuthorDetailsForPost(ctx, feedPost.Post.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author details for post: %w", err)
	}
	feedPost.Post.Author = postAuthor

	for _, comment := range feedPost.Comments {
		commentAuthor, err := fs.getCommentAuthorDetails(ctx, comment.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get author details for comment: %w", err)
		}
		comment.Author = commentAuthor
	}

	return feedPost, nil
}

// getAuthorDetailsForPost is a helper function to retrieve author details for a post.
func (fs *FeedStore) getAuthorDetailsForPost(ctx context.Context, authorID uuid.UUID) (*models.User, error) {
	var author models.User
	author.Role = &models.Role{}
	err := fs.dbPool.QueryRow(ctx, `
		SELECT
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM users u
		INNER JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`, authorID).Scan(
		&author.ID, &author.Username, &author.Email, &author.Banned, &author.IsActive, &author.CreatedAt, &author.UpdatedAt,
		&author.Role.Level, &author.Role.Description,
		&author.Followers, &author.Following,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get author details for post: %w", err)
	}
	return &author, nil
}

// getCommentAuthorDetails is a helper function to retrieve author details for a comment.
func (fs *FeedStore) getCommentAuthorDetails(ctx context.Context, authorID uuid.UUID) (*models.User, error) {
	var author models.User
	author.Role = &models.Role{}
	err := fs.dbPool.QueryRow(ctx, `
		SELECT
			u.id, u.username, u.email, u.banned, u.is_active, u.created_at, u.updated_at,
			r.level, r.description,
			(SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
		FROM users u
		INNER JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`, authorID).Scan(
		&author.ID, &author.Username, &author.Email, &author.Banned, &author.IsActive, &author.CreatedAt, &author.UpdatedAt,
		&author.Role.Level, &author.Role.Description,
		&author.Followers, &author.Following,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get author details for comment: %w", err)
	}
	return &author, nil
}
