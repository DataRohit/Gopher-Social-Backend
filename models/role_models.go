package models

type Role struct {
	Level       int    `json:"level" example:"1"`
	Description string `json:"description" example:"Normal User"`
}
