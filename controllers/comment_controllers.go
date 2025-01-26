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

type CommentController struct {
	commentStore *stores.CommentStore
	postStore    *stores.PostStore
	authStore    *stores.AuthStore
	logger       *logrus.Logger
}

// NewCommentController creates a new CommentController.
//
// Parameters:
//   - commentStore (*stores.CommentStore): CommentStore pointer to interact with the database.
//   - postStore (*stores.PostStore): PostStore pointer to interact with the database.
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *CommentController: Pointer to the CommentController.
func NewCommentController(commentStore *stores.CommentStore, postStore *stores.PostStore, authStore *stores.AuthStore, logger *logrus.Logger) *CommentController {
	return &CommentController{
		commentStore: commentStore,
		postStore:    postStore,
		authStore:    authStore,
		logger:       logger,
	}
}

// CreateComment godoc
// @Summary Create a new comment on a post
// @Description Create a new comment on a post. Requires authentication.
// @Tags comments
// @Accept json
// @Produce json
// @Param postID path string true "Post ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param payload body models.CreateCommentPayload true "Comment payload"
// @Security BearerAuth
// @Success 201 {object} models.CreateCommentSuccessResponse
// @Failure 400 {object} models.CreateCommentErrorResponse
// @Failure 401 {object} models.CreateCommentErrorResponse
// @Failure 500 {object} models.CreateCommentErrorResponse
// @Router /post/{postID}/comment/create [post]
func (cc *CommentController) CreateComment(c *gin.Context) {
	postIDStr := c.Param("postID")
	if postIDStr == "" {
		cc.logger.Error("Post ID is required")
		c.JSON(http.StatusBadRequest, models.CreateCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID is required path parameter",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.CreateCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid postID format",
		})
		return
	}

	var req models.CreateCommentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for Comment Creation")
		c.JSON(http.StatusBadRequest, models.CreateCommentErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		cc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.CreateCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		cc.logger.Error("Invalid user type in context. Middleware misconfiguration.")
		c.JSON(http.StatusInternalServerError, models.CreateCommentErrorResponse{
			Message: "Server Error",
			Error:   "internal server error",
		})
		return
	}

	comment := &models.Comment{
		AuthorID: user.ID,
		PostID:   postID,
		Content:  req.Content,
	}

	createdComment, err := cc.commentStore.CreateComment(c.Request.Context(), comment)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to create comment in store")
		c.JSON(http.StatusInternalServerError, models.CreateCommentErrorResponse{
			Message: "Server Error",
			Error:   "failed to create comment",
		})
		return
	}

	c.JSON(http.StatusCreated, models.CreateCommentSuccessResponse{
		Message: "Comment Created Successfully",
		Comment: createdComment,
	})
}

// UpdateComment godoc
// @Summary Update an existing comment
// @Description Update an existing comment. Requires authentication.
// @Tags comments
// @Accept json
// @Produce json
// @Param postID path string true "Post ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param commentID path string true "Comment ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param payload body models.UpdateCommentPayload true "Comment payload"
// @Security BearerAuth
// @Success 200 {object} models.UpdateCommentSuccessResponse
// @Failure 400 {object} models.UpdateCommentErrorResponse
// @Failure 401 {object} models.UpdateCommentErrorResponse
// @Failure 403 {object} models.UpdateCommentErrorResponse
// @Failure 404 {object} models.UpdateCommentErrorResponse
// @Failure 500 {object} models.UpdateCommentErrorResponse
// @Router /post/{postID}/comment/{commentID}/update [put]
func (cc *CommentController) UpdateComment(c *gin.Context) {
	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		cc.logger.Error("Post ID and Comment ID are required")
		c.JSON(http.StatusBadRequest, models.UpdateCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UpdateCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid postID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.UpdateCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid commentID format",
		})
		return
	}

	var req models.UpdateCommentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for Comment Update")
		c.JSON(http.StatusBadRequest, models.UpdateCommentErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		cc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UpdateCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		cc.logger.Error("Invalid user type in context. Middleware misconfiguration.")
		c.JSON(http.StatusInternalServerError, models.UpdateCommentErrorResponse{
			Message: "Server Error",
			Error:   "internal server error",
		})
		return
	}

	existingComment, err := cc.commentStore.GetCommentByID(c.Request.Context(), commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, models.UpdateCommentErrorResponse{
				Message: "Not Found",
				Error:   "comment not found",
			})
			return
		}
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to get comment from store")
		c.JSON(http.StatusInternalServerError, models.UpdateCommentErrorResponse{
			Message: "Server Error",
			Error:   "failed to get comment",
		})
		return
	}

	if existingComment.AuthorID != user.ID {
		cc.logger.Error("User is not the author of the comment.")
		c.JSON(http.StatusForbidden, models.UpdateCommentErrorResponse{
			Message: "Forbidden",
			Error:   "user is not authorized to update this comment",
		})
		return
	}

	comment := &models.Comment{
		ID:       commentID,
		PostID:   postID,
		AuthorID: user.ID,
		Content:  req.Content,
	}

	updatedComment, err := cc.commentStore.UpdateComment(c.Request.Context(), comment)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to update comment in store")
		c.JSON(http.StatusInternalServerError, models.UpdateCommentErrorResponse{
			Message: "Server Error",
			Error:   "failed to update comment",
		})
		return
	}

	c.JSON(http.StatusOK, models.UpdateCommentSuccessResponse{
		Message: "Comment Updated Successfully",
		Comment: updatedComment,
	})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment. Requires authentication and user must be the author of the comment.
// @Tags comments
// @Accept json
// @Produce json
// @Param postID path string true "Post ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param commentID path string true "Comment ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Security BearerAuth
// @Success 200 {object} models.DeleteCommentSuccessResponse
// @Failure 400 {object} models.DeleteCommentErrorResponse
// @Failure 401 {object} models.DeleteCommentErrorResponse
// @Failure 403 {object} models.DeleteCommentErrorResponse
// @Failure 404 {object} models.DeleteCommentErrorResponse
// @Failure 500 {object} models.DeleteCommentErrorResponse
// @Router /post/{postID}/comment/{commentID}/delete [delete]
func (cc *CommentController) DeleteComment(c *gin.Context) {
	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		cc.logger.Error("Post ID and Comment ID are required")
		c.JSON(http.StatusBadRequest, models.DeleteCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.DeleteCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid postID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.DeleteCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid commentID format",
		})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		cc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DeleteCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		cc.logger.Error("Invalid user type in context. Middleware misconfiguration.")
		c.JSON(http.StatusInternalServerError, models.DeleteCommentErrorResponse{
			Message: "Server Error",
			Error:   "internal server error",
		})
		return
	}

	existingComment, err := cc.commentStore.GetCommentByID(c.Request.Context(), commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, models.DeleteCommentErrorResponse{
				Message: "Not Found",
				Error:   "comment not found",
			})
			return
		}
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to get comment from store")
		c.JSON(http.StatusInternalServerError, models.DeleteCommentErrorResponse{
			Message: "Server Error",
			Error:   "failed to get comment",
		})
		return
	}

	if existingComment.AuthorID != user.ID {
		cc.logger.Error("User is not the author of the comment.")
		c.JSON(http.StatusForbidden, models.DeleteCommentErrorResponse{
			Message: "Forbidden",
			Error:   "user is not authorized to delete this comment",
		})
		return
	}

	err = cc.commentStore.DeleteComment(c.Request.Context(), commentID, postID)
	if err != nil {
		cc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to delete comment from store")
		c.JSON(http.StatusInternalServerError, models.DeleteCommentErrorResponse{
			Message: "Server Error",
			Error:   "failed to delete comment",
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteCommentSuccessResponse{
		Message: "Comment Deleted Successfully",
	})
}
