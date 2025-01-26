package controllers

import (
	"errors"
	"net/http"

	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FeedController struct {
	feedStore *stores.FeedStore
	logger    *logrus.Logger
}

// NewFeedController creates a new FeedController.
//
// Parameters:
//   - feedStore (*stores.FeedStore): FeedStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *FeedController: Pointer to the FeedController.
func NewFeedController(feedStore *stores.FeedStore, logger *logrus.Logger) *FeedController {
	return &FeedController{
		feedStore: feedStore,
		logger:    logger,
	}
}

// ListFeed godoc
// @Summary      List latest posts for feed
// @Description  Retrieves a paginated list of the latest posts for the feed. No authentication required.
// @Tags         feed
// @Accept       json
// @Produce      json
// @Param        page query integer false "Page number for pagination" default(1)
// @Success      200 {object} models.ListFeedSuccessResponse "Successfully retrieved feed posts"
// @Failure      500 {object} models.ListFeedErrorResponse "Internal Server Error - Failed to fetch feed posts"
// @Router       /feed [get]
func (fc *FeedController) ListFeed(c *gin.Context) {
	pageNumber := c.GetInt(middlewares.PageNumberKey)

	posts, err := fc.feedStore.ListLatestPosts(c, pageNumber, middlewares.PageSize)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to get latest posts from store")
		c.JSON(http.StatusInternalServerError, models.ListFeedErrorResponse{
			Message: "Failed to Get Feed Posts",
			Error:   "could not retrieve latest posts from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.ListFeedSuccessResponse{
		Message: "Feed Posts Retrieved Successfully",
		Posts:   posts,
	})
}

// GetFeedPost godoc
// @Summary      Get a specific post with comments for feed
// @Description  Retrieves a specific post by postID along with its comments in paginated form. No authentication required.
// @Tags         feed
// @Accept       json
// @Produce      json
// @Param        postID path string true "Post Identifier (Post ID)"
// @Param        page query integer false "Page number for comments pagination" default(1)
// @Success      200 {object} models.GetFeedPostSuccessResponse "Successfully retrieved feed post with comments"
// @Failure      400 {object} models.GetFeedPostErrorResponse "Bad Request - Invalid Post ID format"
// @Failure      404 {object} models.GetFeedPostErrorResponse "Not Found - Post not found"
// @Failure      500 {object} models.GetFeedPostErrorResponse "Internal Server Error - Failed to fetch feed post with comments"
// @Router       /feed/{postID} [get]
func (fc *FeedController) GetFeedPost(c *gin.Context) {
	pageNumber := c.GetInt(middlewares.PageNumberKey)
	postIDStr := c.Param("postID")
	if postIDStr == "" {
		fc.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.GetFeedPostErrorResponse{
			Message: "Invalid Request",
			Error:   "post ID is required in path",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.GetFeedPostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid post ID format",
		})
		return
	}

	feedPost, err := fc.feedStore.GetPostWithComments(c, postID, pageNumber, middlewares.PageSize)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			fc.logger.WithFields(logrus.Fields{"error": err, "postID": postID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.GetFeedPostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			fc.logger.WithFields(logrus.Fields{"error": err, "postID": postID}).Error("Failed to get post with comments from store")
			c.JSON(http.StatusInternalServerError, models.GetFeedPostErrorResponse{
				Message: "Failed to Get Feed Post",
				Error:   "could not retrieve post with comments from database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.GetFeedPostSuccessResponse{
		Message: "Feed Post with Comments Retrieved Successfully",
		Post:    feedPost,
	})
}
