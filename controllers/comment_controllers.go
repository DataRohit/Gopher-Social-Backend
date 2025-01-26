package controllers

import (
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
// @Tags Comments
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
