package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/gin-gonic/gin"
)

var (
	SERVER_MODE = helpers.GetEnv("SERVER_MODE", "release")
	SERVER_PORT = helpers.GetEnv("SERVER_PORT", ":8080")
)

func main() {
	gin.SetMode(SERVER_MODE)

	router := gin.New()

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
