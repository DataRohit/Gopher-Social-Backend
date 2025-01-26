package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/datarohit/gopher-social-backend/middlewares"
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

// ListTimedOutUsers godoc
// @Summary      List timed out users
// @Description  Retrieves a list of users who are currently timed out. Accessible to moderators and admins.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query integer false "Page number for pagination" default(1)
// @Success      200 {object} models.ListTimedOutUsersSuccessResponse "Successfully retrieved list of timed out users"
// @Failure      401 {object} models.ListTimedOutUsersErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.ListTimedOutUsersErrorResponse "Forbidden - Insufficient permissions"
// @Failure      500 {object} models.ListTimedOutUsersErrorResponse "Internal Server Error - Failed to list timed out users"
// @Router       /action/timeout [get]
func (ac *ActionController) ListTimedOutUsers(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.ListTimedOutUsersErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.ListTimedOutUsersErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	pageNumber := c.GetInt(middlewares.PageNumberKey)

	timedOutUsers, err := ac.actionStore.ListTimedOutUsers(c, pageNumber, middlewares.PageSize)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "requestingUserID": requestingUser.ID}).Error("Failed to list timed out users from store")
		c.JSON(http.StatusInternalServerError, models.ListTimedOutUsersErrorResponse{
			Message: "Failed to List Timed Out Users",
			Error:   "could not retrieve timed out users from database",
		})
		return
	}

	c.JSON(http.StatusOK, models.ListTimedOutUsersSuccessResponse{
		Message: "Timed Out Users Retrieved Successfully",
		Users:   timedOutUsers,
	})
}

// DeactivateUser godoc
// @Summary      Deactivate a user
// @Description  Deactivates a user, preventing them from accessing the platform.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to deactivate"
// @Success      200 {object} models.DeactivateUserSuccessResponse "Successfully deactivated user"
// @Failure      400 {object} models.DeactivateUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.DeactivateUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DeactivateUserErrorResponse "Forbidden - Insufficient permissions or target user cannot be deactivated by requester"
// @Failure      404 {object} models.DeactivateUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.DeactivateUserErrorResponse "Internal Server Error - Failed to deactivate user"
// @Router       /action/deactivate/{userID} [delete]
func (ac *ActionController) DeactivateUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DeactivateUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.DeactivateUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.DeactivateUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.DeactivateUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.DeactivateUserErrorResponse{
				Message: "Failed to Deactivate User",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.DeactivateUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 2 {
		if targetUser.Role.Level > 1 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Moderator cannot deactivate moderator or admin")
			c.JSON(http.StatusForbidden, models.DeactivateUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrModeratorCannotDeactivateModeratorOrAdmin.Error(),
			})
			return
		}
	}

	if requestingUser.Role.Level == 3 {
		if targetUser.Role.Level == 3 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Admin cannot deactivate another admin")
			c.JSON(http.StatusForbidden, models.DeactivateUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrAdminCannotDeactivateAdmin.Error(),
			})
			return
		}
	}

	err = ac.actionStore.DeactivateUser(c, targetUserID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to deactivate user in store")
		c.JSON(http.StatusInternalServerError, models.DeactivateUserErrorResponse{
			Message: "Failed to Deactivate User",
			Error:   "could not deactivate user",
		})
		return
	}

	c.JSON(http.StatusOK, models.DeactivateUserSuccessResponse{
		Message: "User Deactivated Successfully",
	})
}

// ActivateUser godoc
// @Summary      Activate a user
// @Description  Activates a user, allowing them to access the platform.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to activate"
// @Success      200 {object} models.ActivateUserSuccessResponse "Successfully activated user"
// @Failure      400 {object} models.ActivateUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.ActivateUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.ActivateUserErrorResponse "Forbidden - Insufficient permissions or target user cannot be activated by requester"
// @Failure      404 {object} models.ActivateUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.ActivateUserErrorResponse "Internal Server Error - Failed to activate user"
// @Router       /action/activate/{userID} [post]
func (ac *ActionController) ActivateUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.ActivateUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.ActivateUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.ActivateUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.ActivateUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.ActivateUserErrorResponse{
				Message: "Failed to Activate User",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.ActivateUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 2 {
		if targetUser.Role.Level > 1 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Moderator cannot activate moderator or admin")
			c.JSON(http.StatusForbidden, models.ActivateUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrModeratorCannotActivateModeratorOrAdmin.Error(),
			})
			return
		}
	}

	err = ac.actionStore.ActivateUser(c, targetUserID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to activate user in store")
		c.JSON(http.StatusInternalServerError, models.ActivateUserErrorResponse{
			Message: "Failed to Activate User",
			Error:   "could not activate user",
		})
		return
	}

	c.JSON(http.StatusOK, models.ActivateUserSuccessResponse{
		Message: "User Activated Successfully",
	})
}

// UnbanUser godoc
// @Summary      Unban a user
// @Description  Unbans a user, only sets the banned status to false.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to unban"
// @Success      200 {object} models.UnbanUserSuccessResponse "Successfully unbanned user"
// @Failure      400 {object} models.UnbanUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UnbanUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UnbanUserErrorResponse "Forbidden - Insufficient permissions or target user cannot be unbanned by requester"
// @Failure      404 {object} models.UnbanUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.UnbanUserErrorResponse "Internal Server Error - Failed to unban user"
// @Router       /action/unban/{userID} [post]
func (ac *ActionController) UnbanUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.UnbanUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.UnbanUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.UnbanUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.UnbanUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.UnbanUserErrorResponse{
				Message: "Failed to Unban User",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level != 3 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.UnbanUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 3 {
		if targetUser.Role.Level == 3 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Admin cannot unban another admin")
			c.JSON(http.StatusForbidden, models.UnbanUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrAdminCannotUnbanAdmin.Error(),
			})
			return
		}
	}

	err = ac.actionStore.UnbanUser(c, targetUserID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to unban user in store")
		c.JSON(http.StatusInternalServerError, models.UnbanUserErrorResponse{
			Message: "Failed to Unban User",
			Error:   "could not unban user",
		})
		return
	}

	c.JSON(http.StatusOK, models.UnbanUserSuccessResponse{
		Message: "User Unbanned Successfully",
	})
}

// BanUser godoc
// @Summary      Ban a user
// @Description  Bans a user, deactivates them and deletes all their posts.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userID path string true "User ID to ban"
// @Success      200 {object} models.BanUserSuccessResponse "Successfully banned user"
// @Failure      400 {object} models.BanUserErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.BanUserErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.BanUserErrorResponse "Forbidden - Insufficient permissions or target user cannot be banned by requester"
// @Failure      404 {object} models.BanUserErrorResponse "Not Found - User not found"
// @Failure      500 {object} models.BanUserErrorResponse "Internal Server Error - Failed to ban user"
// @Router       /action/ban/{userID} [post]
func (ac *ActionController) BanUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.BanUserErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	targetUserIDStr := c.Param("userID")
	if targetUserIDStr == "" {
		ac.logger.Error("Target User ID is required in path")
		c.JSON(http.StatusBadRequest, models.BanUserErrorResponse{
			Message: "Invalid Request",
			Error:   "target userID is required path parameter",
		})
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": targetUserIDStr}).Error("Invalid Target User ID format")
		c.JSON(http.StatusBadRequest, models.BanUserErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid target userID format",
		})
		return
	}

	targetUser, err := ac.authStore.GetUserByID(c, targetUserID)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Target user not found")
			c.JSON(http.StatusNotFound, models.BanUserErrorResponse{
				Message: "User Not Found",
				Error:   "target user not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to get target user from store")
			c.JSON(http.StatusInternalServerError, models.BanUserErrorResponse{
				Message: "Failed to Ban User",
				Error:   "could not retrieve user details",
			})
		}
		return
	}

	if requestingUser.Role.Level != 3 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.BanUserErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	if requestingUser.Role.Level == 3 {
		if targetUser.Role.Level == 3 {
			ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level, "targetUserID": targetUserID, "targetUserRole": targetUser.Role.Level}).Error("Admin cannot ban another admin")
			c.JSON(http.StatusForbidden, models.BanUserErrorResponse{
				Message: "Forbidden",
				Error:   stores.ErrAdminCannotBanAdmin.Error(),
			})
			return
		}
	}

	err = ac.actionStore.BanUser(c, targetUserID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "targetUserID": targetUserID, "requestingUserID": requestingUser.ID}).Error("Failed to ban user in store")
		c.JSON(http.StatusInternalServerError, models.BanUserErrorResponse{
			Message: "Failed to Ban User",
			Error:   "could not ban user",
		})
		return
	}

	c.JSON(http.StatusOK, models.BanUserSuccessResponse{
		Message: "User Banned Successfully",
	})
}

// DeleteComment godoc
// @Summary      Delete a comment by comment ID
// @Description  Deletes a comment, accessible to moderators and admins.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        commentID path string true "Comment ID to delete"
// @Success      200 {object} models.DeleteCommentSuccessResponse "Successfully deleted comment"
// @Failure      400 {object} models.DeleteCommentErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.DeleteCommentErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DeleteCommentErrorResponse "Forbidden - Insufficient permissions"
// @Failure      404 {object} models.DeleteCommentErrorResponse "Not Found - Comment not found"
// @Failure      500 {object} models.DeleteCommentErrorResponse "Internal Server Error - Failed to delete comment"
// @Router       /action/comment/{commentID} [delete]
func (ac *ActionController) DeleteComment(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DeleteCommentErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	if requestingUser.Role.Level < 2 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.DeleteCommentErrorResponse{
			Message: "Forbidden",
			Error:   "insufficient permissions",
		})
		return
	}

	commentIDStr := c.Param("commentID")
	if commentIDStr == "" {
		ac.logger.Error("Comment ID is required in path")
		c.JSON(http.StatusBadRequest, models.DeleteCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "commentID is required path parameter",
		})
		return
	}

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "commentID": commentIDStr}).Error("Invalid Comment ID format")
		c.JSON(http.StatusBadRequest, models.DeleteCommentErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid commentID format",
		})
		return
	}

	err = ac.actionStore.DeleteCommentByCommentID(c, commentID)
	if err != nil {
		if errors.Is(err, stores.ErrCommentNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "commentID": commentID, "requestingUserID": requestingUser.ID}).Error("Comment not found")
			c.JSON(http.StatusNotFound, models.DeleteCommentErrorResponse{
				Message: "Comment Not Found",
				Error:   "comment not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "commentID": commentID, "requestingUserID": requestingUser.ID}).Error("Failed to delete comment from store")
			c.JSON(http.StatusInternalServerError, models.DeleteCommentErrorResponse{
				Message: "Failed to Delete Comment",
				Error:   "could not delete comment",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.DeleteCommentSuccessResponse{
		Message: "Comment Deleted Successfully",
	})
}

// DeletePost godoc
// @Summary      Delete a post by post ID
// @Description  Deletes a post, accessible to admins only.
// @Tags         action
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID path string true "Post ID to delete"
// @Success      200 {object} models.DeletePostSuccessResponse "Successfully deleted post"
// @Failure      400 {object} models.DeletePostErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.DeletePostErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.DeletePostErrorResponse "Forbidden - Insufficient permissions"
// @Failure      404 {object} models.DeletePostErrorResponse "Not Found - Post not found"
// @Failure      500 {object} models.DeletePostErrorResponse "Internal Server Error - Failed to delete post"
// @Router       /action/post/{postID} [delete]
func (ac *ActionController) DeletePost(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		ac.logger.Error("User not found in context. Middleware misconfiguration.")
		c.JSON(http.StatusUnauthorized, models.DeletePostErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	requestingUser := userCtx.(*models.User)

	if requestingUser.Role.Level != 3 {
		ac.logger.WithFields(logrus.Fields{"requestingUserID": requestingUser.ID, "requestingUserRole": requestingUser.Role.Level}).Error("Unauthorized user role")
		c.JSON(http.StatusForbidden, models.DeletePostErrorResponse{
			Message: "Forbidden",
			Error:   stores.ErrAdminOnlyOperation.Error(),
		})
		return
	}

	postIDStr := c.Param("postID")
	if postIDStr == "" {
		ac.logger.Error("Post ID is required in path")
		c.JSON(http.StatusBadRequest, models.DeletePostErrorResponse{
			Message: "Invalid Request",
			Error:   "postID is required path parameter",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "postID": postIDStr}).Error("Invalid Post ID format")
		c.JSON(http.StatusBadRequest, models.DeletePostErrorResponse{
			Message: "Invalid Request",
			Error:   "invalid postID format",
		})
		return
	}

	err = ac.actionStore.DeletePostByPostID(c, postID)
	if err != nil {
		if errors.Is(err, stores.ErrPostNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "requestingUserID": requestingUser.ID}).Error("Post not found")
			c.JSON(http.StatusNotFound, models.DeletePostErrorResponse{
				Message: "Post Not Found",
				Error:   "post not found",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "postID": postID, "requestingUserID": requestingUser.ID}).Error("Failed to delete post from store")
			c.JSON(http.StatusInternalServerError, models.DeletePostErrorResponse{
				Message: "Failed to Delete Post",
				Error:   "could not delete post",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.DeletePostSuccessResponse{
		Message: "Post Deleted Successfully",
	})
}
