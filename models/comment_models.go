package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AuthorID  uuid.UUID `json:"-" example:"550e8400-e29b-41d4-a716-446655440000"`
	Author    *User     `json:"author,omitempty"`
	PostID    uuid.UUID `json:"-" example:"550e8400-e29b-41d4-a716-446655440000"`
	Post      *Post     `json:"post,omitempty"`
	Content   string    `json:"content" example:"This is a comment content"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-25T12:34:01.159498Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-25T12:34:01.159498Z"`
}

// Create Comment Models
type CreateCommentPayload struct {
	Content string `json:"content" binding:"required,min=1,max=500" example:"This is a comment content"`
}

type CreateCommentSuccessResponse struct {
	Message string   `json:"message" example:"Comment Created Successfully"`
	Comment *Comment `json:"comment"`
}

type CreateCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Update Comment Models
type UpdateCommentPayload struct {
	Content string `json:"content" binding:"required,min=1,max=500" example:"This is a updated comment content"`
}

type UpdateCommentSuccessResponse struct {
	Message string   `json:"message" example:"Comment Updated Successfully"`
	Comment *Comment `json:"comment"`
}

type UpdateCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Delete Comment Models
type DeleteCommentSuccessResponse struct {
	Message string `json:"message" example:"Comment Deleted Successfully"`
}

type DeleteCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Get Comment Models
type GetCommentSuccessResponse struct {
	Message string   `json:"message" example:"Comment Retrieved Successfully"`
	Comment *Comment `json:"comment"`
}

type GetCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List My Comments Models
type ListMyCommentsSuccessResponse struct {
	Message  string     `json:"message" example:"Comments Retrieved Successfully"`
	Comments []*Comment `json:"comments"`
}

type ListMyCommentsErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List User Comments Models
type ListUserCommentsSuccessResponse struct {
	Message  string     `json:"message" example:"Comments Retrieved Successfully"`
	Comments []*Comment `json:"comments"`
}

type ListUserCommentsErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
