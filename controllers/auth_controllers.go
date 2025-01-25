package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	authStore *stores.AuthStore
	logger    *logrus.Logger
}

// NewAuthController creates a new AuthController.
//
// Parameters:
//   - authStore (*stores.AuthStore): AuthStore pointer to interact with the database.
//   - logger (*logrus.Logger): Logrus logger pointer to log messages.
//
// Returns:
//   - *AuthController: Pointer to the AuthController.
func NewAuthController(authStore *stores.AuthStore, logger *logrus.Logger) *AuthController {
	return &AuthController{
		authStore: authStore,
		logger:    logger,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Registers a new user to the platform
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.UserRegisterPayload true "Request Body for User Registration"
// @Success      201 {object} models.UserRegisterSuccessResponse "Successfully registered user"
// @Failure      400 {object} models.UserRegisterErrorResponse "Bad Request - Invalid input"
// @Failure      409 {object} models.UserRegisterErrorResponse "Conflict - User already exists"
// @Failure      500 {object} models.UserRegisterErrorResponse "Internal Server Error - Failed to register user"
// @Router       /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req models.UserRegisterPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for User Registration")
		c.JSON(http.StatusBadRequest, models.UserRegisterErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Hash Password")
		c.JSON(http.StatusInternalServerError, models.UserRegisterErrorResponse{
			Message: "Failed to Register User",
			Error:   "Failed to Hash Password",
		})
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	createdUser, err := ac.authStore.CreateUser(c, user)
	if err != nil {
		if errors.Is(err, stores.ErrUserAlreadyExists) {
			ac.logger.WithFields(logrus.Fields{"error": err, "username": req.Username, "email": req.Email}).Error("User Already Exists")
			c.JSON(http.StatusConflict, models.UserRegisterErrorResponse{
				Message: "User already Exists",
				Error:   err.Error(),
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "username": req.Username, "email": req.Email}).Error("Failed to Create User in Store")
			c.JSON(http.StatusInternalServerError, models.UserRegisterErrorResponse{
				Message: "Failed to Register User",
				Error:   "Failed to Create User",
			})
		}
		return
	}
	log.Printf("User Registered Successfully: %v", createdUser)

	c.JSON(http.StatusCreated, models.UserRegisterSuccessResponse{
		Message: "User Registered Successfully",
		User:    createdUser,
	})
}
