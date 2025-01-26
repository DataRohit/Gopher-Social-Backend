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
type FollowUserSuccessResponse struct {
	Message string `json:"message" example:"User Followed Successfully"`
}

type FollowUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Unfollow User Models
type UnfollowUserSuccessResponse struct {
	Message string `json:"message" example:"User Unfollowed Successfully"`
}

type UnfollowUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
