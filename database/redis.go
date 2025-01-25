package database

import (
	"context"
	"os"
	"strconv"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client

// InitRedis initializes a connection to Redis and stores the client in the RedisClient variable.
// It also checks if the connection was successful and logs the result.
//
// Parameters:
//   - logger (*logrus.Logger): The logger instance to log the result of the connection.
//
// Returns:
//   - None
func InitRedis(logger *logrus.Logger) {
	redisAddr := helpers.GetEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := helpers.GetEnv("REDIS_PASSWORD", "")
	redisDBStr := helpers.GetEnv("REDIS_DB", "0")

	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		redisDB = 0
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Connection to Redis Failed!")
		os.Exit(1)
	}

	logger.Info("Connected to Redis successfully!")
}

// CloseRedis closes the connection to Redis if it is not nil.
// It also logs any errors that occur while closing the connection.
//
// Parameters:
//   - logger (*logrus.Logger): The logger instance to log any errors that occur while closing the connection.
//
// Returns:
//   - None
func CloseRedis(logger *logrus.Logger) {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Error("Error Closing Redis Connection!")
		}
	}
}
