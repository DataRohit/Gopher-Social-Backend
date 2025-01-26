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

// Get Feed Post Models
type GetFeedPostSuccessResponse struct {
	Message string    `json:"message" example:"Feed Post with Comments Retrieved Successfully"`
	Post    *FeedPost `json:"post"`
}

type GetFeedPostErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// FeedPost Model
type FeedPost struct {
	Post     *Post      `json:"post"`
	Comments []*Comment `json:"comments"`
}
