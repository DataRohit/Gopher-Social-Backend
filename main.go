package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/datarohit/gopher-social-backend/database"
	_ "github.com/datarohit/gopher-social-backend/docs"
	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/datarohit/gopher-social-backend/middlewares"
	"github.com/datarohit/gopher-social-backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	SERVER_MODE = helpers.GetEnv("SERVER_MODE", "release")
	SERVER_PORT = helpers.GetEnv("SERVER_PORT", ":8080")
)

// @title           Gopher Social API
// @version         1.0
// @description     This is the API for Gopher Social, a social media platform for Gophers.

// @contact.name   Rohit Vilas Ingole
// @contact.email  rohit.vilas.ingole@gmail.com

// @license.name  MIT License
// @license.url   https://github.com/DataRohit/Gopher-Social-Backend/blob/master/license

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	gin.SetMode(SERVER_MODE)

	logger := helpers.NewLogger()

	database.InitRedis(logger)
	defer database.CloseRedis(logger)

	database.InitPostgres(logger)
	defer database.ClosePostgres(logger)

	router := gin.New()

	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RealIPMiddleware())
	router.Use(middlewares.LoggerMiddleware(logger))
	router.Use(middlewares.RecovererMiddleware(logger))
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.TimeoutMiddleware(10 * time.Second))
	router.Use(middlewares.RateLimiterMiddleware(database.RedisClient, 120, time.Minute, logger))

	apiv1 := router.Group("/api/v1")
	routes.HealthRoutes(apiv1)
	routes.AuthRoutes(apiv1, database.PostgresDB, logger)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:    SERVER_PORT,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithFields(logrus.Fields{"error": err}).Error("Server Failed to Start!")
		}
	}()
	logger.WithFields(logrus.Fields{"mode": SERVER_MODE, "port": SERVER_PORT}).Info("Server Started Successfully!")

	<-quit
	logger.WithFields(logrus.Fields{"signal": "SIGINT"}).Info("Shutdown Signal Received, Exiting Gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Error("Server Forced to Shutdown!")
	}

	logger.WithFields(logrus.Fields{"mode": SERVER_MODE}).Info("Server Shutdown Successfully!")
}
