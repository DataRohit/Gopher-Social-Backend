package controllers

import (
	"errors"
	"net/http"

	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProfileController struct {
	profileStore *stores.ProfileStore
	logger       *logrus.Logger
}

// NewProfileController creates a new ProfileController.
//
// Parameters:
//   - profileStore (*stores.ProfileStore): ProfileStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *ProfileController: Pointer to the ProfileController.
func NewProfileController(profileStore *stores.ProfileStore, logger *logrus.Logger) *ProfileController {
	return &ProfileController{
		profileStore: profileStore,
		logger:       logger,
	}
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the profile of the logged-in user.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.UpdateProfilePayload true "Request Body for Profile Update"
// @Success      200 {object} models.UpdateProfileSuccessResponse "Successfully updated profile"
// @Failure      400 {object} models.UpdateProfileErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UpdateProfileErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.UpdateProfileErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      500 {object} models.UpdateProfileErrorResponse "Internal Server Error - Failed to update profile"
// @Router       /profile/update [put]
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not Found in Context. Middleware Misconfiguration")
		c.JSON(http.StatusUnauthorized, models.UpdateProfileErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := user.(*models.User)

	var req models.UpdateProfilePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Invalid Request Body for Profile Update")
		c.JSON(http.StatusBadRequest, models.UpdateProfileErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	profile := &models.Profile{
		UserID:        userModel.ID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Website:       req.Website,
		Github:        req.Github,
		LinkedIn:      req.LinkedIn,
		Twitter:       req.Twitter,
		GoogleScholar: req.GoogleScholar,
	}

	updatedProfile, err := pc.profileStore.UpdateProfile(c, profile)
	if err != nil {
		if errors.Is(err, stores.ErrProfileNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Profile Not Found for Update")
			c.JSON(http.StatusNotFound, models.UpdateProfileErrorResponse{
				Message: "Profile Not Found",
				Error:   err.Error(),
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to Update Profile in Store")
			c.JSON(http.StatusInternalServerError, models.UpdateProfileErrorResponse{
				Message: "Failed to Update Profile",
				Error:   "failed to update profile in database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.UpdateProfileSuccessResponse{
		Message: "Profile Updated Successfully",
		Profile: updatedProfile,
	})
}

// GetLoggedInUserProfile godoc
// @Summary      Get logged-in user profile
// @Description  Retrieves the profile of the currently logged-in user.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.GetLoggedInUserProfileSuccessResponse "Successfully retrieved profile"
// @Failure      401 {object} models.GetLoggedInUserProfileErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.GetLoggedInUserProfileErrorResponse "Forbidden - User account is inactive or banned"
// @Failure      404 {object} models.GetLoggedInUserProfileErrorResponse "Not Found - Profile not found for the logged-in user"
// @Failure      500 {object} models.GetLoggedInUserProfileErrorResponse "Internal Server Error - Failed to get profile"
// @Router       /profile/me [get]
func (pc *ProfileController) GetLoggedInUserProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not Found in Context. Middleware Misconfiguration")
		c.JSON(http.StatusUnauthorized, models.GetLoggedInUserProfileErrorResponse{
			Message: "Unauthorized",
			Error:   "user not authenticated",
		})
		return
	}
	userModel := user.(*models.User)

	profile, err := pc.profileStore.GetProfileByUserID(c, userModel.ID)
	if err != nil {
		if errors.Is(err, stores.ErrProfileNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Profile Not Found")
			c.JSON(http.StatusNotFound, models.GetLoggedInUserProfileErrorResponse{
				Message: "Profile Not Found",
				Error:   err.Error(),
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "userID": userModel.ID}).Error("Failed to Get Profile from Store")
			c.JSON(http.StatusInternalServerError, models.GetLoggedInUserProfileErrorResponse{
				Message: "Failed to Get Profile",
				Error:   "failed to get profile from database",
			})
		}
		return
	}

	c.JSON(http.StatusOK, models.GetLoggedInUserProfileSuccessResponse{
		Message: "Profile Retrieved Successfully",
		Profile: profile,
	})
}

// GetUserProfile godoc
// @Summary      Get user profile by identifier
// @Description  Retrieves the profile of a user by their identifier (username, email, or user ID).
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        identifier path string true "User Identifier (username, email, or user ID)"
// @Success      200 {object} models.GetUserProfileSuccessResponse "Successfully retrieved profile"
// @Failure      400 {object} models.GetUserProfileErrorResponse "Bad Request - Invalid identifier format"
// @Failure      401 {object} models.GetUserProfileErrorResponse "Unauthorized - User not logged in or invalid token"
// @Failure      403 {object} models.GetUserProfileErrorResponse "Forbidden - User account is inactive or banned or requested user is banned"
// @Failure      404 {object} models.GetUserProfileErrorResponse "Not Found - Profile not found for the given identifier"
// @Failure      500 {object} models.GetUserProfileErrorResponse "Internal Server Error - Failed to get profile"
// @Router       /profile/{identifier} [get]
func (pc *ProfileController) GetUserProfile(c *gin.Context) {
	identifier := c.Param("identifier")

	if identifier == "" {
		pc.logger.Error("Identifier is missing in the request path")
		c.JSON(http.StatusBadRequest, models.GetUserProfileErrorResponse{
			Message: "Invalid Request",
			Error:   "identifier is required in path parameters",
		})
		return
	}

	profile, err := pc.profileStore.GetProfileByIdentifier(c, identifier)
	if err != nil {
		if errors.Is(err, stores.ErrProfileNotFound) {
			pc.logger.WithFields(logrus.Fields{"error": err, "identifier": identifier}).Error("Profile Not Found for Identifier")
			c.JSON(http.StatusNotFound, models.GetUserProfileErrorResponse{
				Message: "Profile Not Found",
				Error:   err.Error(),
			})
		} else {
			pc.logger.WithFields(logrus.Fields{"error": err, "identifier": identifier}).Error("Failed to Get Profile from Store by Identifier")
			c.JSON(http.StatusInternalServerError, models.GetUserProfileErrorResponse{
				Message: "Failed to Get Profile",
				Error:   "failed to get profile from database",
			})
		}
		return
	}

	if profile.User.Banned {
		pc.logger.WithFields(logrus.Fields{"userID": profile.User.ID}).Error("Requested User Profile is Banned")
		c.JSON(http.StatusForbidden, models.GetUserProfileErrorResponse{
			Message: "Forbidden",
			Error:   "requested user profile is banned",
		})
		return
	}
	if !profile.User.IsActive {
		pc.logger.WithFields(logrus.Fields{"userID": profile.User.ID}).Error("Requested User Profile is Inactive")
		c.JSON(http.StatusForbidden, models.GetUserProfileErrorResponse{
			Message: "Forbidden",
			Error:   "requested user profile is inactive",
		})
		return
	}

	c.JSON(http.StatusOK, models.GetUserProfileSuccessResponse{
		Message: "Profile Retrieved Successfully",
		Profile: profile,
	})
}
