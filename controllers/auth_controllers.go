package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/datarohit/gopher-social-backend/models"
	"github.com/datarohit/gopher-social-backend/stores"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DOMAIN is the domain of the application.
var DOMAIN = helpers.GetEnv("DOMAIN", "http://localhost:8080")

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
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Hash Password")
		c.JSON(http.StatusInternalServerError, models.UserRegisterErrorResponse{
			Message: "Failed to Register User",
			Error:   "failed to hash password",
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
				Message: "User Already Exists",
				Error:   err.Error(),
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "username": req.Username, "email": req.Email}).Error("Failed to Create User in Store")
			c.JSON(http.StatusInternalServerError, models.UserRegisterErrorResponse{
				Message: "Failed to Register User",
				Error:   "failed to create user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.UserRegisterSuccessResponse{
		Message: "User Registered Successfully",
		User:    createdUser,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Logs in an existing user and returns access and refresh tokens as secure cookies.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.UserLoginPayload true "Request Body for User Login"
// @Success      200 {object} models.UserLoginSuccessResponse "Successfully logged in"
// @Failure      400 {object} models.UserLoginErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.UserLoginErrorResponse "Unauthorized - Invalid credentials"
// @Failure      500 {object} models.UserLoginErrorResponse "Internal Server Error - Failed to login user"
// @Router       /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	accessTokenCookie, err := c.Cookie("access_token")
	refreshTokenCookie, _ := c.Cookie("refresh_token")

	if err == nil {
		accessToken, err := helpers.VerifyAccessToken(accessTokenCookie)
		if err == nil && accessToken.Valid {
			userID, err := helpers.ExtractUserIDFromToken(accessToken)
			if err != nil {
				ac.logger.WithFields(logrus.Fields{"error": err, "token": accessTokenCookie}).Error("Failed to Extract User ID from Access Token")
				c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
					Message: "Login Failed",
					Error:   "internal server error",
				})
				return
			}

			_, err = ac.authStore.GetUserByID(c, userID)
			if err != nil {
				if errors.Is(err, stores.ErrUserNotFound) {
					goto RefreshOrNormalLogin
				} else {
					ac.logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to Get User by ID")
					c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
						Message: "Login Failed",
						Error:   "internal server error",
					})
					return
				}
			}

			c.JSON(http.StatusOK, models.UserLoginSuccessResponse{
				Message: "User Already Logged In",
			})
			return
		} else {
			if refreshTokenCookie != "" {
				refreshToken, err := helpers.VerifyRefreshToken(refreshTokenCookie)
				if err == nil && refreshToken.Valid {
					userID, err := helpers.ExtractUserIDFromToken(refreshToken)
					if err != nil {
						ac.logger.WithFields(logrus.Fields{"error": err, "token": refreshTokenCookie}).Error("Failed to extract user ID from refresh token")
						c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
							Message: "Login Failed",
							Error:   "internal server error",
						})
						return
					}
					user, err := ac.authStore.GetUserByID(c, userID)
					if err != nil {
						if errors.Is(err, stores.ErrUserNotFound) {
							goto NormalLogin
						} else {
							ac.logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to get user by ID for refresh")
							c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
								Message: "Login Failed",
								Error:   "internal server error",
							})
							return
						}
					}

					newAccessToken, err := helpers.GenerateAccessToken(user.ID)
					if err != nil {
						ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new access token during refresh")
						c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
							Message: "Login Failed",
							Error:   "failed to generate tokens",
						})
						return
					}
					newRefreshToken, err := helpers.GenerateRefreshToken(user.ID)
					if err != nil {
						ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new refresh token during refresh")
						c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
							Message: "Login Failed",
							Error:   "failed to generate tokens",
						})
						return
					}

					c.SetCookie("access_token", newAccessToken, int(time.Minute*30/time.Second), "/", "", true, true)
					c.SetCookie("refresh_token", newRefreshToken, int(time.Hour*6/time.Second), "/", "", true, true)

					log.Printf("User Logged in Successfully (Refreshed Tokens): %v", user.ID)
					c.JSON(http.StatusOK, models.UserLoginSuccessResponse{
						Message: "Login Successful",
					})
					return
				}
			}
		}
	}

RefreshOrNormalLogin:
	if refreshTokenCookie != "" {
		refreshToken, err := helpers.VerifyRefreshToken(refreshTokenCookie)
		if err == nil && refreshToken.Valid {
			userID, err := helpers.ExtractUserIDFromToken(refreshToken)
			if err != nil {
				ac.logger.WithFields(logrus.Fields{"error": err, "token": refreshTokenCookie}).Error("Failed to extract user ID from refresh token")
				c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
					Message: "Login Failed",
					Error:   "internal server error",
				})
				return
			}
			user, err := ac.authStore.GetUserByID(c, userID)
			if err != nil {
				if errors.Is(err, stores.ErrUserNotFound) {
					goto NormalLogin
				} else {
					ac.logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to get user by ID for refresh")
					c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
						Message: "Login Failed",
						Error:   "internal server error",
					})
					return
				}
			}

			newAccessToken, err := helpers.GenerateAccessToken(user.ID)
			if err != nil {
				ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new access token during refresh")
				c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
					Message: "Login Failed",
					Error:   "failed to generate tokens",
				})
				return
			}
			newRefreshToken, err := helpers.GenerateRefreshToken(user.ID)
			if err != nil {
				ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to generate new refresh token during refresh")
				c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
					Message: "Login Failed",
					Error:   "failed to generate tokens",
				})
				return
			}

			c.SetCookie("access_token", newAccessToken, int(time.Minute*30/time.Second), "/", "", true, true)
			c.SetCookie("refresh_token", newRefreshToken, int(time.Hour*6/time.Second), "/", "", true, true)

			log.Printf("User Logged in Successfully (Refreshed Tokens): %v", user.ID)
			c.JSON(http.StatusOK, models.UserLoginSuccessResponse{
				Message: "Login Successful",
			})
			return
		}
	}

NormalLogin:
	var req models.UserLoginPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for User Login")
		c.JSON(http.StatusBadRequest, models.UserLoginErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	user, err := ac.authStore.GetUserByUsernameOrEmail(c, req.Identifier)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"error": err, "identifier": req.Identifier}).Error("User Not Found")
			c.JSON(http.StatusUnauthorized, models.UserLoginErrorResponse{
				Message: "Login Failed",
				Error:   "invalid credentials",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "identifier": req.Identifier}).Error("Failed to Fetch User from Store")
			c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
				Message: "Login Failed",
				Error:   "failed to authenticate user",
			})
		}
		return
	}

	if err := helpers.ComparePassword(user.PasswordHash, req.Password); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "identifier": req.Identifier}).Error("Invalid Password")
		c.JSON(http.StatusUnauthorized, models.UserLoginErrorResponse{
			Message: "Login Failed",
			Error:   "invalid credentials",
		})
		return
	}

	accessToken, err := helpers.GenerateAccessToken(user.ID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to Generate Access Token")
		c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
			Message: "Login Failed",
			Error:   "failed to generate tokens",
		})
		return
	}

	refreshToken, err := helpers.GenerateRefreshToken(user.ID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to Generate Refresh Token")
		c.JSON(http.StatusInternalServerError, models.UserLoginErrorResponse{
			Message: "Login Failed",
			Error:   "failed to generate tokens",
		})
		return
	}

	c.SetCookie("access_token", accessToken, int(time.Minute*30/time.Second), "/", "", true, true)
	c.SetCookie("refresh_token", refreshToken, int(time.Hour*6/time.Second), "/", "", true, true)

	log.Printf("User Logged in Successfully: %v", user.ID)
	c.JSON(http.StatusOK, models.UserLoginSuccessResponse{
		Message: "Login Successful",
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Logs out the current user by clearing access and refresh tokens.
// @Tags         auth
// @Produce      json
// @Success      200 {object} models.UserLogoutSuccessResponse "Successfully logged out"
// @Failure      400 {object} models.UserLogoutErrorResponse "Bad Request - User not logged in"
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	_, errAccessToken := c.Cookie("access_token")
	_, errRefreshToken := c.Cookie("refresh_token")

	if errors.Is(errAccessToken, http.ErrNoCookie) || errors.Is(errRefreshToken, http.ErrNoCookie) {
		ac.logger.WithFields(logrus.Fields{"request-id": c.GetString("request-id")}).Warn("Logout Attempted without Cookies, User Not Logged In")
		c.JSON(http.StatusBadRequest, models.UserLogoutErrorResponse{
			Message: "User Not Logged In",
		})
		return
	}

	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	ac.logger.WithFields(logrus.Fields{"request-id": c.GetString("request-id")}).Info("User Logged Out Successfully")

	c.JSON(http.StatusOK, models.UserLogoutSuccessResponse{
		Message: "Logout Successful",
	})
}

// ForgotPassword godoc
// @Summary      Initiate forgot password flow
// @Description  Initiates the forgot password flow by generating a reset link and sending it to the user's email if the user exists.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.ForgotPasswordPayload true "Request Body for Forgot Password"
// @Success      200 {object} models.ForgotPasswordSuccessResponse "Successfully initiated forgot password flow"
// @Failure      400 {object} models.ForgotPasswordErrorResponse "Bad Request - Invalid input"
// @Failure      500 {object} models.ForgotPasswordErrorResponse "Internal Server Error - Failed to initiate forgot password flow"
// @Router       /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for Forgot Password")
		c.JSON(http.StatusBadRequest, models.ForgotPasswordErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	user, err := ac.authStore.GetUserByUsernameOrEmail(c, req.Identifier)
	if err != nil {
		if errors.Is(err, stores.ErrUserNotFound) {
			ac.logger.WithFields(logrus.Fields{"identifier": req.Identifier}).Info("Forgot Password Request for Non-Existent User")
			c.JSON(http.StatusOK, models.ForgotPasswordSuccessResponse{
				Message: "Password Reset Link Sent Successfully If User Exists",
			})
			return
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "identifier": req.Identifier}).Error("Failed to Get User from Store for Forgot Password")
			c.JSON(http.StatusInternalServerError, models.ForgotPasswordErrorResponse{
				Message: "Failed to Initiate Password Reset",
				Error:   "failed to fetch user",
			})
			return
		}
	}

	resetToken, err := helpers.GeneratePasswordResetToken(user.ID)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to Generate Password Reset Token")
		c.JSON(http.StatusInternalServerError, models.ForgotPasswordErrorResponse{
			Message: "Failed to Initiate Password Reset",
			Error:   "failed to generate reset token",
		})
		return
	}

	expiryTime := helpers.ConvertToAsiaMumbaiTime(time.Now().Add(time.Minute * 15))
	err = ac.authStore.CreatePasswordResetToken(c, user.ID, resetToken, expiryTime)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": user.ID}).Error("Failed to Save Password Reset Token to Store")
		c.JSON(http.StatusInternalServerError, models.ForgotPasswordErrorResponse{
			Message: "Failed to Initiate Password Reset",
			Error:   "failed to save reset token",
		})
		return
	}

	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, models.ForgotPasswordSuccessResponse{
		Message: "Password Reset Link Sent Successfully If User Exists",
		Link:    fmt.Sprintf("%s/api/v1/reset-password?token=%s", DOMAIN, resetToken),
	})
}

// ResetPassword godoc
// @Summary      Reset user password
// @Description  Resets the user's password using the provided reset token in query parameter.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token query string true "Reset Token"
// @Param        body body models.ResetPasswordPayload true "Request Body for Reset Password"
// @Success      200 {object} models.ResetPasswordSuccessResponse "Successfully reset password"
// @Failure      400 {object} models.ResetPasswordErrorResponse "Bad Request - Invalid input"
// @Failure      401 {object} models.ResetPasswordErrorResponse "Unauthorized - Invalid or expired reset token"
// @Failure      500 {object} models.ResetPasswordErrorResponse "Internal Server Error - Failed to reset password"
// @Router       /auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Invalid Request Body for Reset Password")
		c.JSON(http.StatusBadRequest, models.ResetPasswordErrorResponse{
			Message: "Invalid Request Body",
			Error:   err.Error(),
		})
		return
	}

	token := c.Query("token")
	if token == "" {
		ac.logger.WithFields(logrus.Fields{"error": "token missing in query params"}).Error("Invalid Request: Token Missing")
		c.JSON(http.StatusBadRequest, models.ResetPasswordErrorResponse{
			Message: "Invalid Request",
			Error:   "token is required in query parameters",
		})
		return
	}
	// No need to bind token from body anymore
	// req.Token = token

	userID, err := ac.authStore.ValidatePasswordResetToken(c, token, helpers.ConvertToAsiaMumbaiTime(time.Now()))
	if err != nil {
		if errors.Is(err, stores.ErrInvalidOrExpiredToken) {
			ac.logger.WithFields(logrus.Fields{"error": err, "token": token}).Error("Invalid or Expired Reset Token")
			c.JSON(http.StatusUnauthorized, models.ResetPasswordErrorResponse{
				Message: "Invalid or Expired Reset Token",
				Error:   "invalid or expired token",
			})
		} else {
			ac.logger.WithFields(logrus.Fields{"error": err, "token": token}).Error("Failed to Validate Password Reset Token")
			c.JSON(http.StatusInternalServerError, models.ResetPasswordErrorResponse{
				Message: "Failed to Reset Password",
				Error:   "failed to validate reset token",
			})
		}
		return
	}

	hashedPassword, err := helpers.HashPassword(req.NewPassword)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Hash New Password")
		c.JSON(http.StatusInternalServerError, models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password",
			Error:   "failed to hash new password",
		})
		return
	}

	err = ac.authStore.UpdateUserPassword(c, userID, hashedPassword)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "userID": userID}).Error("Failed to Update User Password in Store")
		c.JSON(http.StatusInternalServerError, models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password",
			Error:   "failed to update password",
		})
		return
	}

	err = ac.authStore.InvalidatePasswordResetToken(c, token)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err, "token": token}).Warn("Failed to Invalidate Password Reset Token, But Password Reset was Successful")
	}

	c.JSON(http.StatusOK, models.ResetPasswordSuccessResponse{
		Message: "Password Reset Successfully.",
	})
}
