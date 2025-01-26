package models

type TimeoutDuration string

// Timeout User Models
type TimeoutUserPayload struct {
	TimeoutDuration TimeoutDuration `json:"timeout_duration" binding:"required,oneof=30m 1h 6h 12h 1d" example:"1h"`
}

type TimeoutUserSuccessResponse struct {
	Message string `json:"message" example:"User Timed Out Successfully"`
}

type TimeoutUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Remove Timeout User Models
type RemoveTimeoutUserSuccessResponse struct {
	Message string `json:"message" example:"User Timeout Removed Successfully"`
}

type RemoveTimeoutUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// List Timed Out Users Models
type ListTimedOutUsersSuccessResponse struct {
	Message string  `json:"message" example:"Timed Out Users Retrieved Successfully"`
	Users   []*User `json:"users"`
}

type ListTimedOutUsersErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Deactivate User Models
type DeactivateUserSuccessResponse struct {
	Message string `json:"message" example:"User Deactivated Successfully"`
}

type DeactivateUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Ban User Models
type BanUserSuccessResponse struct {
	Message string `json:"message" example:"User Banned Successfully"`
}

type BanUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Unban User Models
type UnbanUserSuccessResponse struct {
	Message string `json:"message" example:"User Unbanned Successfully"`
}

type UnbanUserErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
