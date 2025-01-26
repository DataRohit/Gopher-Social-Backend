package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ActionController struct {
	authStore   *stores.AuthStore
	actionStore *stores.ActionStore
	logger      *logrus.Logger
}

// NewActionController creates a new ActionController.
//
// Parameters:
//   - actionStore (*stores.ActionStore): ActionStore pointer to interact with user action data.
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with user data.
//   - logger (*logrus.Logger): Logger for logging messages.
//
// Returns:
//   - *ActionController: New ActionController instance.
func NewActionController(actionStore *stores.ActionStore, authStore *stores.AuthStore, logger *logrus.Logger) *ActionController {
	return &ActionController{
		actionStore: actionStore,
		authStore:   authStore,
		logger:      logger,
	}
}

// TimeoutUser godoc
// @Summary      Timeout a user
// @Description  Applies a timeout to a user, restricting their access for a specified duration.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to timeout"
// @Param        body body models.TimeoutUserPayload true "Request Body for timeout duration"
// @Success      200 {object} models.TimeoutUserSuccessResponse "Successfully timed out user"
// @Failure      400 {object} models.TimeoutUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.TimeoutUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.TimeoutUserErrorResponse "Forbidden - Insufficient permissions or target user cannot be timed out by requester"
// @Failure      404 {object} models.TimeoutUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.TimeoutUserErrorResponse "Internal Server Error - Failed to timeout user"
// @Router       /action/timeout/{userID} [post]
func (ac *ActionController) TimeoutUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.TimeoutUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.TimeoutUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.TimeoutUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.TimeoutUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.TimeoutUserErrorResponse{
				Message: "Failed to Timeout User",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.TimeoutUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 2 {
		if targetUser.Role.Level > 1 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Moderator cannot timeout moderator or admin")
			c.JSON(http.StatusForbidden, models.TimeoutUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrModeratorCannotTimeoutModeratorOrAdmin.Error(),
			})
			return
		}
	}

	if requestingUser.Role.Level == 3 {
		if targetUser.Role.Level == 3 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Admin cannot timeout another admin")
			c.JSON(http.StatusForbidden, models.TimeoutUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrAdminCannotTimeoutAdmin.Error(),
			})
			return
		}
	}

	var req models.TimeoutUserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Invalid request body for timeout user")
		c.JSON(http.StatusBadRequest, models.TimeoutUserErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	var timeoutDuration time.Duration
	switch strings.ToLower(string(req.TimeoutDuration)) {
	case "30m":
		timeoutDuration = 30 * time.Minute
	case "1h":
		timeoutDuration = time.Hour
	case "6h":
		timeoutDuration = 6 * time.Hour
	case "12h":
		timeoutDuration = 12 * time.Hour
	case "1d":
		timeoutDuration = 24 * time.Hour
	default:
		ac.logger.WithFields(logrus.Fields{"duration": req.TimeoutDuration, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Invalid timeout duration provided")
		c.JSON(http.StatusBadRequest, models.TimeoutUserErrorResponse{
			Message: "Invalid Timeout Duration",
			Error:   "timeout duration must be one of: 30m, 1h, 6h, 12h, 1d",
		})
		return
	}

	err = ac.actionStore.TimeoutUser(c, targetUserID, timeoutDuration)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID, "duration": req.TimeoutDuration}).Error("Failed to timeout user in store")
		c.JSON(http.StatusInternalServerError, models.TimeoutUserErrorResponse{
			Message: "Failed to Timeout User",
			Error:   "could not apply timeout to user",
		})
		return
	}

	c.JSON(http.StatusOK, models.TimeoutUserSuccessResponse{
		Message: "User Timed Out Successfully",
	})
}

// RemoveTimeoutUser godoc
// @Summary      Remove timeout from a user
// @Description  Removes an active timeout from a user, restoring their access.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to remove timeout from"
// @Success      200 {object} models.RemoveTimeoutUserSuccessResponse "Successfully removed user timeout"
// @Failure      400 {object} models.RemoveTimeoutUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.RemoveTimeoutUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.RemoveTimeoutUserErrorResponse "Forbidden - Insufficient permissions or target user cannot have timeout removed by requester"
// @Failure      404 {object} models.RemoveTimeoutUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.RemoveTimeoutUserErrorResponse "Internal Server Error - Failed to remove user timeout"
// @Router       /action/timeout/{userID} [delete]
func (ac *ActionController) RemoveTimeoutUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.RemoveTimeoutUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.RemoveTimeoutUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.RemoveTimeoutUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.RemoveTimeoutUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.RemoveTimeoutUserErrorResponse{
				Message: "Failed to Remove User Timeout",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.RemoveTimeoutUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 2 {
		if targetUser.Role.Level > 1 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Moderator cannot remove timeout from moderator or admin")
			c.JSON(http.StatusForbidden, models.RemoveTimeoutUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrModeratorCannotTimeoutModeratorOrAdmin.Error(),
			})
			return
		}
	}

	if requestingUser.Role.Level == 3 {
		if targetUser.Role.Level == 3 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Admin cannot remove timeout from another admin")
			c.JSON(http.StatusForbidden, models.RemoveTimeoutUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrAdminCannotTimeoutAdmin.Error(),
			})
			return
		}
	}

	err = ac.actionStore.RemoveTimeoutUser(c, targetUserID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to remove timeout from user in store")
		c.JSON(http.StatusInternalServerError, models.RemoveTimeoutUserErrorResponse{
			Message: "Failed to Remove User Timeout",
			Error:   "could not remove timeout from user",
		})
		return
	}

	c.JSON(http.StatusOK, models.RemoveTimeoutUserSuccessResponse{
		Message: "User Timeout Removed Successfully",
	})
}
