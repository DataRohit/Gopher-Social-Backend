package models

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Level       int       `json:"level" example:"1"`
	Description string    `json:"description" example:"Normal User"`
}
