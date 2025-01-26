package controllers

import (
	"net/http"

	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
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
