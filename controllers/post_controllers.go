package controllers

import (
	"net/http"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PostController handles post related requests.
type PostController struct {
	postStore *stores.PostStore
	authStore *stores.AuthStore
	logger    *logrus.Logger
}

// NewPostController creates a new PostController.
//
// Parameters:
//   - postStore (*stores.PostStore): PostStore pointer to interact with the database.
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *PostController: Pointer to the PostController.
func NewPostController(postStore *stores.PostStore, authStore *stores.AuthStore, logger *logrus.Logger) *PostController {
	return &PostController{
		postStore: postStore,
		authStore: authStore,
		logger:    logger,
	}
}

// CreatePost godoc
// @Summary      Create a new post
// @Description  Creates a new post by a logged-in user.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.CreatePostPayload true "Request Body for creating a post"
// @Success      201 {object} models.CreatePostSuccessResponse "Successfully created post"
// @Failure      400 {object} models.CreatePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.CreatePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.CreatePostErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      500 {object} models.CreatePostErrorResponse "Internal Server Error - Failed to create post"
// @Router       /post/create [post]
func (pc *PostController) CreatePost(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.CreatePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := user.(*models.User)

	var req models.CreatePostPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Invalid request body for creating post")
		c.JSON(http.StatusBadRequest, models.CreatePostErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	post := &models.Post{
		AuthorID:    userModel.ID,
		Title:       req.Title,
		SubTitle:    req.SubTitle,
		Description: req.Description,
		Content:     req.Content,
	}

	createdPost, err := pc.postStore.CreatePost(c, post)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to create post in store")
		c.JSON(http.StatusInternalServerError, models.CreatePostErrorResponse{
			Message: "Failed to Create Post",
			Error:   "could not save post to database",
		})
		return
	}

	author, err := pc.authStore.GetUserByID(c, createdPost.AuthorID)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": createdPost.ID, "authorID": createdPost.AuthorID}).Error("Failed to fetch author details")
		c.JSON(http.StatusInternalServerError, models.CreatePostErrorResponse{
			Message: "Failed to Create Post",
			Error:   "could not fetch author details",
		})
		return
	}
	createdPost.Author = author

	c.JSON(http.StatusCreated, models.CreatePostSuccessResponse{
		Message: "Post Created Successfully",
		Post:    createdPost,
	})
}
