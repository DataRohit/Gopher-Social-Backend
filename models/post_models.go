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
