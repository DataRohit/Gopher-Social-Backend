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

type CommentLikesController struct {
	commentLikesStore *stores.CommentLikeStore
	commentStore      *stores.CommentStore
	postStore         *stores.PostStore
	authStore         *stores.AuthStore
	logger            *logrus.Logger
}

// NewCommentLikesController creates a new CommentLikesController.
//
// Parameters:
//   - commentLikesStore (*stores.CommentLikeStore): CommentLikeStore pointer to interact with the database.
//   - commentStore (*stores.CommentStore): CommentStore pointer to interact with the database.
//   - postStore (*stores.PostStore): PostStore pointer to interact with the database.
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *CommentLikesController: Pointer to the CommentLikesController.
func NewCommentLikesController(commentLikesStore *stores.CommentLikeStore, commentStore *stores.CommentStore, postStore *stores.PostStore, authStore *stores.AuthStore, logger *logrus.Logger) *CommentLikesController {
	return &CommentLikesController{
		commentLikesStore: commentLikesStore,
		commentStore:      commentStore,
		postStore:         postStore,
		authStore:         authStore,
		logger:            logger,
	}
}

// LikeComment godoc
// @Summary      Like a comment
// @Description  Allows a logged-in user to like a comment by comment identifier (commentID) under a post (postID).
// @Tags         comment_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID    path     string  true  "Post Identifier (Post ID)"
// @Param        commentID path     string  true  "Comment Identifier (Comment ID)"
// @Success      200 {object} models.LikeCommentSuccessResponse "Successfully liked comment"
// @Failure      400 {object} models.LikeCommentErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.LikeCommentErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.LikeCommentErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.LikeCommentErrorResponse "Not Found - Post or Comment not found"
// @Failure      409 {object} models.LikeCommentErrorResponse "Conflict - Already liked comment"
// @Failure      500 {object} models.LikeCommentErrorResponse "Internal Server Error - Failed to like comment"
// @Router       /post/{postID}/comment/{commentID}/like [post]
func (clc *CommentLikesController) LikeComment(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		clc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.LikeCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		clc.logger.Error("Post ID and Comment ID are required in path")
		c.JSON(http.StatusBadRequest, models.LikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.LikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "commentID": commentIDStr}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.LikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid comment ID format",
		})
		return
	}

	_, err = clc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.LikeCommentErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.LikeCommentErrorResponse{
				Message: "Failed to Like Comment",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = clc.commentStore.GetCommentByID(c, commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment not found")
			c.JSON(http.StatusNotFound, models.LikeCommentErrorResponse{
				Message: "Comment Not Found",
				Error:   "comment not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get comment from store")
			c.JSON(http.StatusInternalServerError, models.LikeCommentErrorResponse{
				Message: "Failed to Like Comment",
				Error:   "could not retrieve comment from database",
			})
		}
		return
	}

	_, err = clc.commentLikesStore.LikeComment(c, userModel.ID, commentID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentLikeAlreadyExists) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment Like Already Exists")
			c.JSON(http.StatusConflict, models.LikeCommentErrorResponse{
				Message: "Like Comment Failed",
				Error:   "already liked comment",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to Like Comment in Store")
			c.JSON(http.StatusInternalServerError, models.LikeCommentErrorResponse{
				Message: "Failed to Like Comment",
				Error:   "could not like comment in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.LikeCommentSuccessResponse{
		Message: "Comment Liked Successfully",
	})
}

// UnlikeComment godoc
// @Summary      Unlike a comment
// @Description  Allows a logged-in user to unlike a comment by comment identifier (commentID) under a post (postID).
// @Tags         comment_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID    path     string  true  "Post Identifier (Post ID)"
// @Param        commentID path     string  true  "Comment Identifier (Comment ID)"
// @Success      200 {object} models.UnlikeCommentSuccessResponse "Successfully unliked comment"
// @Failure      400 {object} models.UnlikeCommentErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UnlikeCommentErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UnlikeCommentErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UnlikeCommentErrorResponse "Not Found - Post or Comment not found or like not found"
// @Failure      500 {object} models.UnlikeCommentErrorResponse "Internal Server Error - Failed to unlike comment"
// @Router       /post/{postID}/comment/{commentID}/like [delete]
func (clc *CommentLikesController) UnlikeComment(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		clc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UnlikeCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		clc.logger.Error("Post ID and Comment ID are required in path")
		c.JSON(http.StatusBadRequest, models.UnlikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UnlikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "commentID": commentIDStr}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.UnlikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid comment ID format",
		})
		return
	}

	_, err = clc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.UnlikeCommentErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.UnlikeCommentErrorResponse{
				Message: "Failed to Unlike Comment",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = clc.commentStore.GetCommentByID(c, commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment not found")
			c.JSON(http.StatusNotFound, models.UnlikeCommentErrorResponse{
				Message: "Comment Not Found",
				Error:   "comment not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get comment from store")
			c.JSON(http.StatusInternalServerError, models.UnlikeCommentErrorResponse{
				Message: "Failed to Unlike Comment",
				Error:   "could not retrieve comment from database",
			})
		}
		return
	}

	err = clc.commentLikesStore.UnlikeComment(c, userModel.ID, commentID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentLikeNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment Like Not Found")
			c.JSON(http.StatusNotFound, models.UnlikeCommentErrorResponse{
				Message: "Unlike Comment Failed",
				Error:   "comment like not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to Unlike Comment in Store")
			c.JSON(http.StatusInternalServerError, models.UnlikeCommentErrorResponse{
				Message: "Failed to Unlike Comment",
				Error:   "could not unlike comment in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UnlikeCommentSuccessResponse{
		Message: "Comment Unliked Successfully",
	})
}

// DislikeComment godoc
// @Summary      Dislike a comment
// @Description  Allows a logged-in user to dislike a comment by comment identifier (commentID) under a post (postID).
// @Tags         comment_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID    path     string  true  "Post Identifier (Post ID)"
// @Param        commentID path     string  true  "Comment Identifier (Comment ID)"
// @Success      200 {object} models.DislikeCommentSuccessResponse "Successfully disliked comment"
// @Failure      400 {object} models.DislikeCommentErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.DislikeCommentErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DislikeCommentErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.DislikeCommentErrorResponse "Not Found - Post or Comment not found"
// @Failure      409 {object} models.DislikeCommentErrorResponse "Conflict - Already disliked comment"
// @Failure      500 {object} models.DislikeCommentErrorResponse "Internal Server Error - Failed to dislike comment"
// @Router       /post/{postID}/comment/{commentID}/dislike [post]
func (clc *CommentLikesController) DislikeComment(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		clc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DislikeCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		clc.logger.Error("Post ID and Comment ID are required in path")
		c.JSON(http.StatusBadRequest, models.DislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.DislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "commentID": commentIDStr}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.DislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid comment ID format",
		})
		return
	}

	_, err = clc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.DislikeCommentErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.DislikeCommentErrorResponse{
				Message: "Failed to Dislike Comment",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = clc.commentStore.GetCommentByID(c, commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment not found")
			c.JSON(http.StatusNotFound, models.DislikeCommentErrorResponse{
				Message: "Comment Not Found",
				Error:   "comment not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get comment from store")
			c.JSON(http.StatusInternalServerError, models.DislikeCommentErrorResponse{
				Message: "Failed to Dislike Comment",
				Error:   "could not retrieve comment from database",
			})
		}
		return
	}

	_, err = clc.commentLikesStore.DislikeComment(c, userModel.ID, commentID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentDislikeAlreadyExists) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment Dislike Already Exists")
			c.JSON(http.StatusConflict, models.DislikeCommentErrorResponse{
				Message: "Dislike Comment Failed",
				Error:   "already disliked comment",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to Dislike Comment in Store")
			c.JSON(http.StatusInternalServerError, models.DislikeCommentErrorResponse{
				Message: "Failed to Dislike Comment",
				Error:   "could not dislike comment in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.DislikeCommentSuccessResponse{
		Message: "Comment Disliked Successfully",
	})
}

// UndislikeComment godoc
// @Summary      Undislike a comment
// @Description  Allows a logged-in user to remove dislike from a comment by comment identifier (commentID) under a post (postID).
// @Tags         comment_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID    path     string  true  "Post Identifier (Post ID)"
// @Param        commentID path     string  true  "Comment Identifier (Comment ID)"
// @Success      200 {object} models.UndislikeCommentSuccessResponse "Successfully removed dislike from comment"
// @Failure      400 {object} models.UndislikeCommentErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UndislikeCommentErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UndislikeCommentErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UndislikeCommentErrorResponse "Not Found - Post or Comment not found or dislike not found"
// @Failure      500 {object} models.UndislikeCommentErrorResponse "Internal Server Error - Failed to remove dislike from comment"
// @Router       /post/{postID}/comment/{commentID}/dislike [delete]
func (clc *CommentLikesController) UndislikeComment(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		clc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UndislikeCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	commentIDStr := c.Param("commentID")

	if postIDStr == "" || commentIDStr == "" {
		clc.logger.Error("Post ID and Comment ID are required in path")
		c.JSON(http.StatusBadRequest, models.UndislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "postID and commentID are required path parameters",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UndislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		clc.logger.WithFields(logrus.Fields{"error": err, "commentID": commentIDStr}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.UndislikeCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid comment ID format",
		})
		return
	}

	_, err = clc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.UndislikeCommentErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.UndislikeCommentErrorResponse{
				Message: "Failed to Undislike Comment",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = clc.commentStore.GetCommentByID(c, commentID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment not found")
			c.JSON(http.StatusNotFound, models.UndislikeCommentErrorResponse{
				Message: "Comment Not Found",
				Error:   "comment not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to get comment from store")
			c.JSON(http.StatusInternalServerError, models.UndislikeCommentErrorResponse{
				Message: "Failed to Undislike Comment",
				Error:   "could not retrieve comment from database",
			})
		}
		return
	}

	err = clc.commentLikesStore.UndislikeComment(c, userModel.ID, commentID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentDislikeNotFound) {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Comment Dislike Not Found")
			c.JSON(http.StatusNotFound, models.UndislikeCommentErrorResponse{
				Message: "Undislike Comment Failed",
				Error:   "comment dislike not found",
			})
		} else {
			clc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "commentID": commentID, "userID": userModel.ID}).Error("Failed to Undislike Comment in Store")
			c.JSON(http.StatusInternalServerError, models.UndislikeCommentErrorResponse{
				Message: "Failed to Undislike Comment",
				Error:   "could not undislike comment in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UndislikeCommentSuccessResponse{
		Message: "Comment Undisliked Successfully",
	})
}
