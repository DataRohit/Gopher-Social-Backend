package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AuthorID  uuid.UUID `json:"author_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Author    *User     `json:"author,omitempty"`
	PostID    uuid.UUID `json:"post_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Post      *Post     `json:"post,omitempty"`
	Content   string    `json:"content" example:"This is a comment content"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-25T12:34:01.159498Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-25T12:34:01.159498Z"`
}
