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
