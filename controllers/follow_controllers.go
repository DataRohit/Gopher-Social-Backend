package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FollowController struct {
	authStore    *stores.AuthStore
	profileStore *stores.ProfileStore
	followStore  *stores.FollowStore
	logger       *logrus.Logger
}

// NewFollowController creates a new FollowController.
//
// Parameters:
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - profileStore (*stores.ProfileStore): ProfileStore pointer to interact with the database.
//   - followStore (*stores.FollowStore): FollowStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *FollowController: Pointer to the FollowController.
func NewFollowController(authStore *stores.AuthStore, profileStore *stores.ProfileStore, followStore *stores.FollowStore, logger *logrus.Logger) *FollowController {
	return &FollowController{
		authStore:    authStore,
		profileStore: profileStore,
		followStore:  followStore,
		logger:       logger,
	}
}

// FollowUser godoc
// @Summary      Follow a user
// @Description  Allows a logged-in user to follow another user.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.FollowUserPayload true "Request Body for Follow User"
// @Success      200 {object} models.FollowUserSuccessResponse "Successfully followed user"
// @Failure      400 {object} models.FollowUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.FollowUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.FollowUserErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.FollowUserErrorResponse "Not Found - Followee user not found"
// @Failure      409 {object} models.FollowUserErrorResponse "Conflict - Already following user"
// @Failure      500 {object} models.FollowUserErrorResponse "Internal Server Error - Failed to follow user"
// @Router       /user/follow [post]
func (fc *FollowController) FollowUser(c *gin.Context) {
	followerUser, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not Found in Context. Middleware Misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.FollowUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	followerUserModel := followerUser.(*models.User)

	var req models.FollowUserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID}).Error("Invalid Request Body for Follow User")
		c.JSON(http.StatusBadRequest, models.FollowUserErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	if req.Identifier == "" {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID}).Error("Followee Identifier is required")
		c.JSON(http.StatusBadRequest, models.FollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "followee identifier is required",
		})
		return
	}

	var followeeUserID uuid.UUID
	parsedUUID, err := uuid.Parse(req.Identifier)
	if err == nil {
		followeeUserID = parsedUUID
	} else {
		followeeUser, err := fc.profileStore.GetProfileByUserID(context.Background(), followerUserModel.ID)
		if err == nil {
			if req.Identifier == followeeUser.User.Username || req.Identifier == followeeUser.User.Email {
				followeeUserID = followeeUser.User.ID
			}
		}
		if followeeUserID == uuid.Nil {
			followeeUserFromAuth, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), req.Identifier)
			if err != nil {
				fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "identifier": req.Identifier}).Error("Followee User Not Found")
				c.JSON(http.StatusNotFound, models.FollowUserErrorResponse{
					Message: "Follow User Failed",
					Error:   "followee user not found",
				})
				return
			}
			followeeUserID = followeeUserFromAuth.ID
		}
	}

	if followeeUserID == followerUserModel.ID {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Cannot follow yourself")
		c.JSON(http.StatusBadRequest, models.FollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "cannot follow yourself",
		})
		return
	}

	err = fc.followStore.FollowUser(context.Background(), followerUserModel.ID, followeeUserID)
	if err != nil {
		if errors.Is(err, stores.ErrAlreadyFollowing) {
			fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Already Following User")
			c.JSON(http.StatusConflict, models.FollowUserErrorResponse{
				Message: "Follow User Failed",
				Error:   "already following user",
			})
		} else {
			fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Failed to Follow User")
			c.JSON(http.StatusInternalServerError, models.FollowUserErrorResponse{
				Message: "Failed to Follow User",
				Error:   "failed to follow user in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.FollowUserSuccessResponse{
		Message: "User Followed Successfully",
	})
}

// UnfollowUser godoc
// @Summary      Unfollow a user
// @Description  Allows a logged-in user to unfollow another user.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.UnfollowUserPayload true "Request Body for Unfollow User"
// @Success      200 {object} models.UnfollowUserSuccessResponse "Successfully unfollowed user"
// @Failure      400 {object} models.UnfollowUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UnfollowUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UnfollowUserErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UnfollowUserErrorResponse "Not Found - Followee user not found or not following"
// @Failure      500 {object} models.UnfollowUserErrorResponse "Internal Server Error - Failed to unfollow user"
// @Router       /user/unfollow [delete]
func (fc *FollowController) UnfollowUser(c *gin.Context) {
	followerUser, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not Found in Context. Middleware Misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UnfollowUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	followerUserModel := followerUser.(*models.User)

	var req models.UnfollowUserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID}).Error("Invalid Request Body for Unfollow User")
		c.JSON(http.StatusBadRequest, models.UnfollowUserErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	if req.Identifier == "" {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID}).Error("Followee Identifier is required")
		c.JSON(http.StatusBadRequest, models.UnfollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "followee identifier is required",
		})
		return
	}

	var followeeUserID uuid.UUID
	parsedUUID, err := uuid.Parse(req.Identifier)
	if err == nil {
		followeeUserID = parsedUUID
	} else {
		followeeUser, err := fc.profileStore.GetProfileByUserID(context.Background(), followerUserModel.ID)
		if err == nil {
			if req.Identifier == followeeUser.User.Username || req.Identifier == followeeUser.User.Email {
				followeeUserID = followeeUser.User.ID
			}
		}
		if followeeUserID == uuid.Nil {
			followeeUserFromAuth, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), req.Identifier)
			if err != nil {
				fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "identifier": req.Identifier}).Error("Followee User Not Found")
				c.JSON(http.StatusNotFound, models.UnfollowUserErrorResponse{
					Message: "Unfollow User Failed",
					Error:   "followee user not found",
				})
				return
			}
			followeeUserID = followeeUserFromAuth.ID
		}
	}

	if followeeUserID == followerUserModel.ID {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Cannot unfollow yourself")
		c.JSON(http.StatusBadRequest, models.UnfollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "cannot unfollow yourself",
		})
		return
	}

	err = fc.followStore.UnfollowUser(context.Background(), followerUserModel.ID, followeeUserID)
	if err != nil {
		if errors.Is(err, stores.ErrNotFollowing) {
			fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Not Following User")
			c.JSON(http.StatusNotFound, models.UnfollowUserErrorResponse{
				Message: "Unfollow User Failed",
				Error:   "not following user",
			})
		} else {
			fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "followeeUserID": followeeUserID}).Error("Failed to Unfollow User")
			c.JSON(http.StatusInternalServerError, models.UnfollowUserErrorResponse{
				Message: "Failed to Unfollow User",
				Error:   "failed to unfollow user in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UnfollowUserSuccessResponse{
		Message: "User Unfollowed Successfully",
	})
}
