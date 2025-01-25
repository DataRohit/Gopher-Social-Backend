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
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	pc.logger.Error("UserID not Found in Context. Middleware Misconfiguration.")
	// 	c.JSON(http.StatusUnauthorized, models.UpdateProfileErrorResponse{
	// 		Message: "Unauthorized",
	// 		Error:   "user not authenticated",
	// 	})
	// 	return
	// }

	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("User not Found in Context. Middleware Misconfiguration.")
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
