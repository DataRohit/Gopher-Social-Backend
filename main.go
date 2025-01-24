package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/datarohit/gopher-social-backend/docs"
	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/gin-gonic/gin"
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

	router := gin.New()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:    SERVER_PORT,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server Failed to Start: %v", err)
		}
	}()
	log.Printf("Server is Running on Port %s", SERVER_PORT)

	<-quit
	log.Printf("Shutdown Signal Received, Exiting Gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Forced to Shutdown: %v", err)
	}

	log.Printf("Server Exited Cleanly!")
}
