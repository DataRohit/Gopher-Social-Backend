package models

import (
	"time"

	"github.com/google/uuid"
)

type CommentLike struct {
	UserID    uuid.UUID `json:"user_id"`
	CommentID uuid.UUID `json:"comment_id"`
	Liked     bool      `json:"liked"`
	CreatedAt time.Time `json:"created_at"`
}

// Like Comment Models
type LikeCommentSuccessResponse struct {
	Message string `json:"message" example:"Comment Liked Successfully"`
}

type LikeCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Unlike Comment Models
type UnlikeCommentSuccessResponse struct {
	Message string `json:"message" example:"Comment Unliked Successfully"`
}

type UnlikeCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Dislike Comment Models
type DislikeCommentSuccessResponse struct {
	Message string `json:"message" example:"Comment Disliked Successfully"`
}

type DislikeCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Undislike Comment Models
type UndislikeCommentSuccessResponse struct {
	Message string `json:"message" example:"Comment Undisliked Successfully"`
}

type UndislikeCommentErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List Liked Comments Under Post Models
type ListLikedCommentsUnderPostSuccessResponse struct {
	Message  string     `json:"message" example:"Liked Comments Retrieved Successfully"`
	Comments []*Comment `json:"comments"`
}

type ListLikedCommentsUnderPostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List Disliked Comments Under Post Models
type ListDislikedCommentsUnderPostSuccessResponse struct {
	Message  string     `json:"message" example:"Disliked Comments Retrieved Successfully"`
	Comments []*Comment `json:"comments"`
}

type ListDislikedCommentsUnderPostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
