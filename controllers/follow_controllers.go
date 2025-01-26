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
// @Param        identifier path string true "User Identifier (username, email, or user ID) of the followee"
// @Success      200 {object} models.FollowUserSuccessResponse "Successfully followed user"
// @Failure      400 {object} models.FollowUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.FollowUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.FollowUserErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.FollowUserErrorResponse "Not Found - Followee user not found"
// @Failure      409 {object} models.FollowUserErrorResponse "Conflict - Already following user"
// @Failure      500 {object} models.FollowUserErrorResponse "Internal Server Error - Failed to follow user"
// @Router       /user/follow/{identifier} [post]
func (fc *FollowController) FollowUser(c *gin.Context) {
	followerUser, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not Found in Context. Middleware Misconfiguration")
		c.JSON(http.StatusUnauthorized, models.FollowUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	followerUserModel := followerUser.(*models.User)

	identifier := c.Param("identifier")
	if identifier == "" {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID}).Error("Followee Identifier is required")
		c.JSON(http.StatusBadRequest, models.FollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "followee identifier is required in path",
		})
		return
	}

	var followeeUserID uuid.UUID
	parsedUUID, err := uuid.Parse(identifier)
	if err == nil {
		followeeUserID = parsedUUID
	} else {
		followeeUser, err := fc.profileStore.GetProfileByUserID(context.Background(), followerUserModel.ID)
		if err == nil {
			if identifier == followeeUser.User.Username || identifier == followeeUser.User.Email {
				followeeUserID = followeeUser.User.ID
			}
		}
		if followeeUserID == uuid.Nil {
			followeeUserFromAuth, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), identifier)
			if err != nil {
				fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "identifier": identifier}).Error("Followee User Not Found")
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
// @Param        identifier path string true "User Identifier (username, email, or user ID) of the followee"
// @Success      200 {object} models.UnfollowUserSuccessResponse "Successfully unfollowed user"
// @Failure      400 {object} models.UnfollowUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UnfollowUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UnfollowUserErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.UnfollowUserErrorResponse "Not Found - Followee user not found or not following"
// @Failure      500 {object} models.UnfollowUserErrorResponse "Internal Server Error - Failed to unfollow user"
// @Router       /user/unfollow/{identifier} [delete]
func (fc *FollowController) UnfollowUser(c *gin.Context) {
	followerUser, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not Found in Context. Middleware Misconfiguration")
		c.JSON(http.StatusUnauthorized, models.UnfollowUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	followerUserModel := followerUser.(*models.User)

	identifier := c.Param("identifier")
	if identifier == "" {
		fc.logger.WithFields(logrus.Fields{"followerUserID": followerUserModel.ID}).Error("Followee Identifier is required")
		c.JSON(http.StatusBadRequest, models.UnfollowUserErrorResponse{
			Message: "Invalid Request",
			Error:   "followee identifier is required in path",
		})
		return
	}

	var followeeUserID uuid.UUID
	parsedUUID, err := uuid.Parse(identifier)
	if err == nil {
		followeeUserID = parsedUUID
	} else {
		followeeUser, err := fc.profileStore.GetProfileByUserID(context.Background(), followerUserModel.ID)
		if err == nil {
			if identifier == followeeUser.User.Username || identifier == followeeUser.User.Email {
				followeeUserID = followeeUser.User.ID
			}
		}
		if followeeUserID == uuid.Nil {
			followeeUserFromAuth, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), identifier)
			if err != nil {
				fc.logger.WithFields(logrus.Fields{"error": err, "followerUserID": followerUserModel.ID, "identifier": identifier}).Error("Followee User Not Found")
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

// GetFollowers godoc
// @Summary      Get followers of logged-in user
// @Description  Retrieves a list of users who are following the logged-in user.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.GetFollowersSuccessResponse "Successfully retrieved followers list"
// @Failure      401 {object} models.GetFollowersErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      500 {object} models.GetFollowersErrorResponse "Internal Server Error - Failed to fetch followers"
// @Router       /user/followers [get]
func (fc *FollowController) GetFollowers(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.GetFollowersErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	followers, err := fc.followStore.GetFollowersByUserID(context.Background(), userModel.ID)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to get followers")
		c.JSON(http.StatusInternalServerError, models.GetFollowersErrorResponse{
			Message: "Failed to Get Followers",
			Error:   "could not retrieve followers from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetFollowersSuccessResponse{
		Message:   "Followers Retrieved Successfully",
		Followers: followers,
	})
}

// GetFollowing godoc
// @Summary      Get users being followed by logged-in user
// @Description  Retrieves a list of users that the logged-in user is following.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.GetFollowingSuccessResponse "Successfully retrieved following list"
// @Failure      401 {object} models.GetFollowingErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      500 {object} models.GetFollowingErrorResponse "Internal Server Error - Failed to fetch following users"
// @Router       /user/following [get]
func (fc *FollowController) GetFollowing(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		fc.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.GetFollowingErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := userCtx.(*models.User)

	following, err := fc.followStore.GetFollowingByUserID(context.Background(), userModel.ID)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to get following users")
		c.JSON(http.StatusInternalServerError, models.GetFollowingErrorResponse{
			Message: "Failed to Get Following Users",
			Error:   "could not retrieve following users from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetFollowingSuccessResponse{
		Message:   "Following Users Retrieved Successfully",
		Following: following,
	})
}

// GetUserFollowers godoc
// @Summary      Get followers of a user by identifier
// @Description  Retrieves a list of users who are following the user identified by identifier.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        identifier path string true "User Identifier (username, email, or user ID) of the user"
// @Success      200 {object} models.GetUserFollowersSuccessResponse "Successfully retrieved followers list for user"
// @Failure      400 {object} models.GetUserFollowersErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.GetUserFollowersErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      404 {object} models.GetUserFollowersErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.GetUserFollowersErrorResponse "Internal Server Error - Failed to fetch followers"
// @Router       /user/{identifier}/followers [get]
func (fc *FollowController) GetUserFollowers(c *gin.Context) {
	identifier := c.Param("identifier")
	if identifier == "" {
		fc.logger.Error("User Identifier is required")
		c.JSON(http.StatusBadRequest, models.GetUserFollowersErrorResponse{
			Message: "Invalid Request",
			Error:   "user identifier is required in path",
		})
		return
	}

	var requestedUserID uuid.UUID
	parsedUUID, err := uuid.Parse(identifier)
	if err == nil {
		requestedUserID = parsedUUID
	} else {
		requestedUser, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), identifier)
		if err != nil {
			fc.logger.WithFields(logrus.Fields{"error": err, "identifier": identifier}).Error("User Not Found")
			c.JSON(http.StatusNotFound, models.GetUserFollowersErrorResponse{
				Message: "Get User Followers Failed",
				Error:   "user not found",
			})
			return
		}
		requestedUserID = requestedUser.ID
	}

	followers, err := fc.followStore.GetFollowersByUserID(context.Background(), requestedUserID)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "userID": requestedUserID}).Error("Failed to get followers for user")
		c.JSON(http.StatusInternalServerError, models.GetUserFollowersErrorResponse{
			Message: "Failed to Get User Followers",
			Error:   "could not retrieve followers from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetUserFollowersSuccessResponse{
		Message:   "User Followers Retrieved Successfully",
		Followers: followers,
	})
}

// GetUserFollowing godoc
// @Summary      Get users being followed by a user by identifier
// @Description  Retrieves a list of users that the user identified by identifier is following.
// @Tags         follow
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        identifier path string true "User Identifier (username, email, or user ID) of the user"
// @Success      200 {object} models.GetUserFollowingSuccessResponse "Successfully retrieved following list for user"
// @Failure      400 {object} models.GetUserFollowingErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.GetUserFollowingErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      404 {object} models.GetUserFollowingErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.GetUserFollowingErrorResponse "Internal Server Error - Failed to fetch following users"
// @Router       /user/{identifier}/following [get]
func (fc *FollowController) GetUserFollowing(c *gin.Context) {
	identifier := c.Param("identifier")
	if identifier == "" {
		fc.logger.Error("User Identifier is required")
		c.JSON(http.StatusBadRequest, models.GetUserFollowingErrorResponse{
			Message: "Invalid Request",
			Error:   "user identifier is required in path",
		})
		return
	}

	var requestedUserID uuid.UUID
	parsedUUID, err := uuid.Parse(identifier)
	if err == nil {
		requestedUserID = parsedUUID
	} else {
		requestedUser, err := fc.authStore.GetUserByUsernameOrEmail(context.Background(), identifier)
		if err != nil {
			fc.logger.WithFields(logrus.Fields{"error": err, "identifier": identifier}).Error("User Not Found")
			c.JSON(http.StatusNotFound, models.GetUserFollowingErrorResponse{
				Message: "Get User Following Failed",
				Error:   "user not found",
			})
			return
		}
		requestedUserID = requestedUser.ID
	}

	following, err := fc.followStore.GetFollowingByUserID(context.Background(), requestedUserID)
	if err != nil {
		fc.logger.WithFields(logrus.Fields{"error": err, "userID": requestedUserID}).Error("Failed to get following users for user")
		c.JSON(http.StatusInternalServerError, models.GetUserFollowingErrorResponse{
			Message: "Failed to Get User Following Users",
			Error:   "could not retrieve following users from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetUserFollowingSuccessResponse{
		Message:   "User Following Users Retrieved Successfully",
		Following: following,
	})
}
