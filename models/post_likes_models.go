package models

import (
	"time"

	"github.com/google/uuid"
)

type PostLike struct {
	UserID    uuid.UUID `json:"user_id"`
	PostID    uuid.UUID `json:"post_id"`
	Liked     bool      `json:"liked"`
	CreatedAt time.Time `json:"created_at"`
}

// Like Post Models
type LikePostPayload struct {
	PostID string `json:"post_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type LikePostSuccessResponse struct {
	Message string `json:"message" example:"Post Liked Successfully"`
}

type LikePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Dislike Post Models
type DislikePostPayload struct {
	PostID string `json:"post_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type DislikePostSuccessResponse struct {
	Message string `json:"message" example:"Post Disliked Successfully"`
}

type DislikePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Unlike Post Models
type UnlikePostSuccessResponse struct {
	Message string `json:"message" example:"Post Unliked Successfully"`
}

type UnlikePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Undislike Post Models
type UndislikePostSuccessResponse struct {
	Message string `json:"message" example:"Post Undisliked Successfully"`
}

type UndislikePostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
