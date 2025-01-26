package models

// List Feed Models
type ListFeedSuccessResponse struct {
	Message string  `json:"message" example:"Feed Posts Retrieved Successfully"`
	Posts   []*Post `json:"posts"`
}

type ListFeedErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
