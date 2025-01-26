package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AuthorID    uuid.UUID `json:"-"`
	Author      *User     `json:"author,omitempty"`
	Title       string    `json:"title" example:"My Awesome Post"`
	SubTitle    string    `json:"sub_title,omitempty" example:"A Catchy Subtitle"`
	Description string    `json:"description,omitempty" example:"A brief description of the post."`
	Content     string    `json:"content" example:"This is the main content of my post."`
	Likes       uint      `json:"likes" example:"100"`
	Dislikes    uint      `json:"dislikes" example:"10"`
	CreatedAt   time.Time `json:"created_at" example:"2025-01-25T12:34:01.159498Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-01-25T12:34:01.159498Z"`
}

// Create Post Models
type CreatePostPayload struct {
	Title       string `json:"title" binding:"required,min=3,max=255" example:"My Awesome Post"`
	SubTitle    string `json:"sub_title,omitempty" example:"A Catchy Subtitle"`
	Description string `json:"description,omitempty" example:"A brief description of the post."`
	Content     string `json:"content" binding:"required" example:"This is the main content of my post."`
}

type CreatePostSuccessResponse struct {
	Message string `json:"message" example:"Post Created Successfully"`
	Post    *Post  `json:"post"`
}

type CreatePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Get Post Models
type GetPostSuccessResponse struct {
	Message string `json:"message" example:"Post Retrieved Successfully"`
	Post    *Post  `json:"post"`
}

type GetPostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Update Post Models
type UpdatePostPayload struct {
	Title       string `json:"title,omitempty" example:"Updated Awesome Post"`
	SubTitle    string `json:"sub_title,omitempty" example:"Updated Catchy Subtitle"`
	Description string `json:"description,omitempty" example:"Updated brief description of the post."`
	Content     string `json:"content,omitempty" example:"Updated main content of my post."`
}

type UpdatePostSuccessResponse struct {
	Message string `json:"message" example:"Post Updated Successfully"`
	Post    *Post  `json:"post"`
}

type UpdatePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Delete Post Models
type DeletePostSuccessResponse struct {
	Message string `json:"message" example:"Post Deleted Successfully"`
}

type DeletePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List My Posts Models
type ListMyPostsSuccessResponse struct {
	Message string  `json:"message" example:"User Posts Retrieved Successfully"`
	Posts   []*Post `json:"posts"`
}

type ListMyPostsErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
