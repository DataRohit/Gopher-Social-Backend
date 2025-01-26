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
