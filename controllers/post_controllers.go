package controllers

import (
	"errors"
	"net/http"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// UpdatePost godoc
// @Summary      Update an existing post
// @Description  Updates an existing post by its ID. Only the author can update the post.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post ID to be updated"
// @Param        body body models.UpdatePostPayload true "Request Body for updating a post"
// @Success      200 {object} models.UpdatePostSuccessResponse "Successfully updated post"
// @Failure      400 {object} models.UpdatePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UpdatePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UpdatePostErrorResponse "Forbidden - User is not the author or account is inactive/banned"
// @Failure      404 {object} models.UpdatePostErrorResponse "Not Found - Post not found"
// @Failure      500 {object} models.UpdatePostErrorResponse "Internal Server Error - Failed to update post"
// @Router       /post/{postID} [put]
func (pc *PostController) UpdatePost(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UpdatePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := user.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		pc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.UpdatePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UpdatePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	var req models.UpdatePostPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Invalid request body for updating post")
		c.JSON(http.StatusBadRequest, models.UpdatePostErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	existingPost, err := pc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.UpdatePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.UpdatePostErrorResponse{
				Message: "Failed to Update Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	if existingPost.AuthorID != userModel.ID {
		pc.logger.WithFields(logrus.Fields{"postID": postID, "userID": userModel.ID, "authorID": existingPost.AuthorID}).Error("User is not the author of the post")
		c.JSON(http.StatusForbidden, models.UpdatePostErrorResponse{
			Message: "Forbidden",
			Error:   "you are not the author of this post",
		})
		return
	}

	post := &models.Post{
		ID:          postID,
		Title:       req.Title,
		SubTitle:    req.SubTitle,
		Description: req.Description,
		Content:     req.Content,
	}

	updatedPost, err := pc.postStore.UpdatePost(c, post)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to update post in store")
		c.JSON(http.StatusInternalServerError, models.UpdatePostErrorResponse{
			Message: "Failed to Update Post",
			Error:   "could not update post in database",
		})
		return
	}

	author, err := pc.authStore.GetUserByID(c, updatedPost.AuthorID)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": updatedPost.ID, "authorID": updatedPost.AuthorID}).Error("Failed to fetch author details after update")
		c.JSON(http.StatusInternalServerError, models.UpdatePostErrorResponse{
			Message: "Failed to Update Post",
			Error:   "could not fetch author details after update",
		})
		return
	}
	updatedPost.Author = author

	c.JSON(http.StatusOK, models.UpdatePostSuccessResponse{
		Message: "Post Updated Successfully",
		Post:    updatedPost,
	})
}

// DeletePost godoc
// @Summary      Delete an existing post
// @Description  Deletes an existing post by its ID. Only the author can delete the post.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post ID to be deleted"
// @Success      200 {object} models.DeletePostSuccessResponse "Successfully deleted post"
// @Failure      401 {object} models.DeletePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DeletePostErrorResponse "Forbidden - User is not the author or account is inactive/banned"
// @Failure      404 {object} models.DeletePostErrorResponse "Not Found - Post not found"
// @Failure      500 {object} models.DeletePostErrorResponse "Internal Server Error - Failed to delete post"
// @Router       /post/{postID} [delete]
func (pc *PostController) DeletePost(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DeletePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := user.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		pc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.DeletePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.DeletePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	existingPost, err := pc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.DeletePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.DeletePostErrorResponse{
				Message: "Failed to Delete Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	if existingPost.AuthorID != userModel.ID {
		pc.logger.WithFields(logrus.Fields{"postID": postID, "userID": userModel.ID, "authorID": existingPost.AuthorID}).Error("User is not the author of the post")
		c.JSON(http.StatusForbidden, models.DeletePostErrorResponse{
			Message: "Forbidden",
			Error:   "you are not the author of this post",
		})
		return
	}

	err = pc.postStore.DeletePost(c, postID)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to delete post from store")
		c.JSON(http.StatusInternalServerError, models.DeletePostErrorResponse{
			Message: "Failed to Delete Post",
			Error:   "could not delete post from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.DeletePostSuccessResponse{
		Message: "Post Deleted Successfully",
	})
}

// GetPost godoc
// @Summary      Get a post by ID
// @Description  Retrieves a post by its ID. Any logged-in user can access this route.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post ID to be retrieved"
// @Success      200 {object} models.GetPostSuccessResponse "Successfully retrieved post"
// @Failure      400 {object} models.GetPostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.GetPostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      404 {object} models.GetPostErrorResponse "Not Found - Post not found"
// @Failure      500 {object} models.GetPostErrorResponse "Internal Server Error - Failed to get post"
// @Router       /post/{postID} [get]
func (pc *PostController) GetPost(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.GetPostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		pc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.GetPostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.GetPostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	retrievedPost, err := pc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.GetPostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "postID": postID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.GetPostErrorResponse{
				Message: "Failed to Get Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	author, err := pc.authStore.GetUserByID(c, retrievedPost.AuthorID)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "postID": retrievedPost.ID, "authorID": retrievedPost.AuthorID}).Error("Failed to fetch author details")
		c.JSON(http.StatusInternalServerError, models.GetPostErrorResponse{
			Message: "Failed to Get Post",
			Error:   "could not fetch author details",
		})
		return
	}
	retrievedPost.Author = author

	c.JSON(http.StatusOK, models.GetPostSuccessResponse{
		Message: "Post Retrieved Successfully",
		Post:    retrievedPost,
	})
}
