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

// Get Followers Models
type GetFollowersSuccessResponse struct {
	Message   string  `json:"message" example:"Followers Retrieved Successfully"`
	Followers []*User `json:"followers"`
}

type GetFollowersErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Get Following Models
type GetFollowingSuccessResponse struct {
	Message   string  `json:"message" example:"Following Users Retrieved Successfully"`
	Following []*User `json:"following"`
}

type GetFollowingErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Get User Followers Models
type GetUserFollowersSuccessResponse struct {
	Message   string  `json:"message" example:"User Followers Retrieved Successfully"`
	Followers []*User `json:"followers"`
}

type GetUserFollowersErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Get User Following Models
type GetUserFollowingSuccessResponse struct {
	Message   string  `json:"message" example:"User Following Users Retrieved Successfully"`
	Following []*User `json:"following"`
}

type GetUserFollowingErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
