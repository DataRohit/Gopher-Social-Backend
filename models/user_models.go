package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username           string     `json:"username" example:"john_doe"`
	Email              string     `json:"email" example:"john.doe@example.com"`
	PasswordHash       string     `json:"-"`
	TimeoutUntil       *time.Time `json:"timeout_until,omitempty" example:"2024-03-15T10:00:00+05:30"`
	Banned             bool       `json:"banned" example:"false"`
	Followers          []*User    `json:"followers,omitempty"`
	Following          []*User    `json:"following,omitempty"`
	CreatedAt          time.Time  `json:"created_at" example:"2024-01-25T07:00:00+05:30"`
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-01-25T07:00:00+05:30"`
	PasswordResetToken *string    `json:"-"`
	ResetTokenExpiry   *time.Time `json:"-"`
}

// User Register Models
type UserRegisterPayload struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"P@$$wOrd"`
}
type UserRegisterSuccessResponse struct {
	Message string `json:"message" example:"User Registered Successfully"`
	User    *User  `json:"user"`
}

type UserRegisterErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// User Login Models
type UserLoginPayload struct {
	Identifier string `json:"identifier" binding:"required" example:"john_doe / john.doe@example.com"`
	Password   string `json:"password" binding:"required,min=8,max=64" example:"P@$$wOrd"`
}

type UserLoginSuccessResponse struct {
	Message string `json:"message" example:"User Logged In Successfully"`
}

type UserLoginErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// User Logout Models
type UserLogoutSuccessResponse struct {
	Message string `json:"message" example:"Logout Successful"`
}

type UserLogoutErrorResponse struct {
	Message string `json:"message" example:"User Not Logged In"`
}

// User Forgot Password Models
type ForgotPasswordPayload struct {
	Identifier string `json:"identifier" binding:"required" example:"john_doe / john.doe@example.com"`
}

type ForgotPasswordSuccessResponse struct {
	Message string `json:"message" example:"Password Reset Link Sent Successfully If User Exists."`
	Link    string `json:"link" example:"http://localhost:8080/api/v1/reset-password?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzc4MDI5ODEsInVzZXJfaWQiOiI1MzAzODI0OS02Yjk4LTQ2YzUtOWQ1YS00ZDdkYjY5MmJiOGMifQ.pxrhavurRWfBlgAYShPnFl7rVcaJn8TsDHM-ZtcuAVg"`
}

type ForgotPasswordErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// User Reset Password Models
type ResetPasswordPayload struct {
	NewPassword     string `json:"new_password" binding:"required,min=8,max=64" example:"NewP@$$wOrd"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword" example:"NewP@$$wOrd"`
}

type ResetPasswordSuccessResponse struct {
	Message string `json:"message" example:"Password Reset Successfully."`
}

type ResetPasswordErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
