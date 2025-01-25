package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID            uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID        uuid.UUID `json:"-"`
	User          *User     `json:"user,omitempty"`
	FirstName     string    `json:"first_name,omitempty" example:"John"`
	LastName      string    `json:"last_name,omitempty" example:"Doe"`
	Website       string    `json:"website,omitempty" example:"https://example.com"`
	Github        string    `json:"github,omitempty" example:"https://github.com/john_doe"`
	LinkedIn      string    `json:"linkedin,omitempty" example:"https://linkedin.com/in/john_doe"`
	Twitter       string    `json:"twitter,omitempty" example:"https://twitter.com/john_doe"`
	GoogleScholar string    `json:"google_scholar,omitempty" example:"https://scholar.google.com/citations?user=xxxxxxxxxxxxx"`
	CreatedAt     time.Time `json:"created_at" example:"2025-01-25T12:34:01.159498Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2025-01-25T12:34:01.159498Z"`
}

// Update Profile Models
type UpdateProfilePayload struct {
	FirstName     string `json:"first_name,omitempty" example:"John"`
	LastName      string `json:"last_name,omitempty" example:"Doe"`
	Website       string `json:"website,omitempty" example:"https://example.com"`
	Github        string `json:"github,omitempty" example:"https://github.com/john_doe"`
	LinkedIn      string `json:"linkedin,omitempty" example:"https://linkedin.com/in/john_doe"`
	Twitter       string `json:"twitter,omitempty" example:"https://twitter.com/john_doe"`
	GoogleScholar string `json:"google_scholar,omitempty" example:"https://scholar.google.com/citations?user=xxxxxxxxxxxxx"`
}

type UpdateProfileSuccessResponse struct {
	Message string   `json:"message" example:"Profile Updated Successfully"`
	Profile *Profile `json:"profile"`
}

type UpdateProfileErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
