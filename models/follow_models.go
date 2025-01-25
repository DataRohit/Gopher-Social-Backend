package models

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID uuid.UUID `json:"-"`
	Follower   *User     `json:"follower"`
	FolloweeID uuid.UUID `json:"-"`
	Followee   *User     `json:"followee"`
	CreatedAt  time.Time `json:"created_at"`
}

// Follow User Models
type FollowUserPayload struct {
	Identifier string `json:"identifier" binding:"required" example:"john_doe / john.doe@example.com / 550e8400-e29b-41d4-a716-446655440000"`
}

type FollowUserSuccessResponse struct {
	Message string `json:"message" example:"User Followed Successfully"`
}

type FollowUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Unfollow User Models
type UnfollowUserPayload struct {
	Identifier string `json:"identifier" binding:"required" example:"john_doe / john.doe@example.com / 550e8400-e29b-41d4-a716-446655440000"`
}

type UnfollowUserSuccessResponse struct {
	Message string `json:"message" example:"User Unfollowed Successfully"`
}

type UnfollowUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
