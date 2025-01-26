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

// PostLikesController handles post like related requests.
type PostLikesController struct {
	postLikesStore *stores.PostLikeStore
	postStore      *stores.PostStore
	authStore      *stores.AuthStore
	logger         *logrus.Logger
}

// NewPostLikesController creates a new PostLikesController.
//
// Parameters:
//   - postLikesStore (*stores.PostLikeStore): PostLikeStore pointer to interact with the database.
//   - postStore (*stores.PostStore): PostStore pointer to interact with the database.
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *PostLikesController: Pointer to the PostLikesController.
func NewPostLikesController(postLikesStore *stores.PostLikeStore, postStore *stores.PostStore, authStore *stores.AuthStore, logger *logrus.Logger) *PostLikesController {
	return &PostLikesController{
		postLikesStore: postLikesStore,
		postStore:      postStore,
		authStore:      authStore,
		logger:         logger,
	}
}

// LikePost godoc
// @Summary      Like a post
// @Description  Allows a logged-in user to like a post by post identifier (postID).
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post Identifier (Post ID)"
// @Success      200 {object} models.LikePostSuccessResponse "Successfully liked post"
// @Failure      400 {object} models.LikePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.LikePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.LikePostErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.LikePostErrorResponse "Not Found - Post not found"
// @Failure      409 {object} models.LikePostErrorResponse "Conflict - Already liked post"
// @Failure      500 {object} models.LikePostErrorResponse "Internal Server Error - Failed to like post"
// @Router       /post/{postID}/like [post]
func (plc *PostLikesController) LikePost(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.LikePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		plc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.LikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.LikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	_, err = plc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.LikePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.LikePostErrorResponse{
				Message: "Failed to Like Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = plc.postLikesStore.LikePost(c, userModel.ID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostLikeAlreadyExists) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post Like Already Exists")
			c.JSON(http.StatusConflict, models.LikePostErrorResponse{
				Message: "Like Post Failed",
				Error:   "already liked post",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to Like Post in Store")
			c.JSON(http.StatusInternalServerError, models.LikePostErrorResponse{
				Message: "Failed to Like Post",
				Error:   "could not like post in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.LikePostSuccessResponse{
		Message: "Post Liked Successfully",
	})
}

// DislikePost godoc
// @Summary      Dislike a post
// @Description  Allows a logged-in user to dislike a post by post identifier (postID).
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post Identifier (Post ID)"
// @Success      200 {object} models.DislikePostSuccessResponse "Successfully disliked post"
// @Failure      400 {object} models.DislikePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.DislikePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DislikePostErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.DislikePostErrorResponse "Not Found - Post not found"
// @Failure      409 {object} models.DislikePostErrorResponse "Conflict - Already disliked post"
// @Failure      500 {object} models.DislikePostErrorResponse "Internal Server Error - Failed to dislike post"
// @Router       /post/{postID}/dislike [post]
func (plc *PostLikesController) DislikePost(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DislikePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		plc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.DislikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.DislikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	_, err = plc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.DislikePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.DislikePostErrorResponse{
				Message: "Failed to Dislike Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	_, err = plc.postLikesStore.DislikePost(c, userModel.ID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostDislikeAlreadyExists) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post Dislike Already Exists")
			c.JSON(http.StatusConflict, models.DislikePostErrorResponse{
				Message: "Dislike Post Failed",
				Error:   "already disliked post",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to Dislike Post in Store")
			c.JSON(http.StatusInternalServerError, models.DislikePostErrorResponse{
				Message: "Failed to Dislike Post",
				Error:   "could not dislike post in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.DislikePostSuccessResponse{
		Message: "Post Disliked Successfully",
	})
}

// UnlikePost godoc
// @Summary      Unlike a post
// @Description  Allows a logged-in user to unlike a post by post identifier (postID).
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post Identifier (Post ID)"
// @Success      200 {object} models.UnlikePostSuccessResponse "Successfully unliked post"
// @Failure      400 {object} models.UnlikePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UnlikePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UnlikePostErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UnlikePostErrorResponse "Not Found - Post not found or like not found"
// @Failure      500 {object} models.UnlikePostErrorResponse "Internal Server Error - Failed to unlike post"
// @Router       /post/{postID}/like [delete]
func (plc *PostLikesController) UnlikePost(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UnlikePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		plc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.UnlikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UnlikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	_, err = plc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.UnlikePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.UnlikePostErrorResponse{
				Message: "Failed to Unlike Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	err = plc.postLikesStore.UnlikePost(c, userModel.ID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostLikeNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post Like Not Found")
			c.JSON(http.StatusNotFound, models.UnlikePostErrorResponse{
				Message: "Unlike Post Failed",
				Error:   "post like not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to Unlike Post in Store")
			c.JSON(http.StatusInternalServerError, models.UnlikePostErrorResponse{
				Message: "Failed to Unlike Post",
				Error:   "could not unlike post in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UnlikePostSuccessResponse{
		Message: "Post Unliked Successfully",
	})
}

// UndislikePost godoc
// @Summary      Undislike a post
// @Description  Allows a logged-in user to undislike a post by post identifier (postID).
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post Identifier (Post ID)"
// @Success      200 {object} models.UndislikePostSuccessResponse "Successfully undisliked post"
// @Failure      400 {object} models.UndislikePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UndislikePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UndislikePostErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UndislikePostErrorResponse "Not Found - Post not found or dislike not found"
// @Failure      500 {object} models.UndislikePostErrorResponse "Internal Server Error - Failed to undislike post"
// @Router       /post/{postID}/undislike [delete]
func (plc *PostLikesController) UndislikePost(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UndislikePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		plc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.UndislikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.UndislikePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	_, err = plc.postStore.GetPostByID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.UndislikePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to get post from store")
			c.JSON(http.StatusInternalServerError, models.UndislikePostErrorResponse{
				Message: "Failed to Undislike Post",
				Error:   "could not retrieve post from database",
			})
		}
		return
	}

	err = plc.postLikesStore.UndislikePost(c, userModel.ID, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostDislikeNotFound) {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Post Dislike Not Found")
			c.JSON(http.StatusNotFound, models.UndislikePostErrorResponse{
				Message: "Undislike Post Failed",
				Error:   "post dislike not found",
			})
		} else {
			plc.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "userID": userModel.ID}).Error("Failed to Undislike Post in Store")
			c.JSON(http.StatusInternalServerError, models.UndislikePostErrorResponse{
				Message: "Failed to Undislike Post",
				Error:   "could not undislike post in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UndislikePostSuccessResponse{
		Message: "Post Undisliked Successfully",
	})
}

// ListLikedPosts godoc
// @Summary      List liked posts of logged-in user
// @Description  Retrieves a list of posts liked by the logged-in user.
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.ListLikedPostsSuccessResponse "Successfully retrieved list of liked posts"
// @Failure      401 {object} models.ListLikedPostsErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      500 {object} models.ListLikedPostsErrorResponse "Internal Server Error - Failed to fetch liked posts"
// @Router       /post/liked [get]
func (plc *PostLikesController) ListLikedPosts(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.ListLikedPostsErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	posts, err := plc.postLikesStore.ListLikedPostsByUserID(c, userModel.ID)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to get liked posts from store")
		c.JSON(http.StatusInternalServerError, models.ListLikedPostsErrorResponse{
			Message: "Failed to Get Liked Posts",
			Error:   "could not retrieve liked posts from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.ListLikedPostsSuccessResponse{
		Message: "Liked Posts Retrieved Successfully",
		Posts:   posts,
	})
}

// ListDislikedPosts godoc
// @Summary      List disliked posts of logged-in user
// @Description  Retrieves a list of posts disliked by the logged-in user.
// @Tags         post_likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.ListDislikedPostsSuccessResponse "Successfully retrieved list of disliked posts"
// @Failure      401 {object} models.ListDislikedPostsErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      500 {object} models.ListDislikedPostsErrorResponse "Internal Server Error - Failed to fetch disliked posts"
// @Router       /post/disliked [get]
func (plc *PostLikesController) ListDislikedPosts(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		plc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.ListDislikedPostsErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	posts, err := plc.postLikesStore.ListDislikedPostsByUserID(c, userModel.ID)
	if err != nil {
		plc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to get disliked posts from store")
		c.JSON(http.StatusInternalServerError, models.ListDislikedPostsErrorResponse{
			Message: "Failed to Get Disliked Posts",
			Error:   "could not retrieve disliked posts from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.ListDislikedPostsSuccessResponse{
		Message: "Disliked Posts Retrieved Successfully",
		Posts:   posts,
	})
}
