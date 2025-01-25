package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RateLimiterMiddleware is a middleware that limits the number of requests from a single IP address.
// It uses Redis to store the request counts and enforces a rate limit of 'limit' requests per 'duration'.
// If the client exceeds the rate limit, the middleware responds with a 429 Too Many Requests error.
//
// Parameters:
//   - redisClient (*redis.Client): Redis client to use for rate limiting.
//   - limit int: Maximum number of requests allowed within the duration.
//   - duration time.Duration: Time window for the rate limit.
//   - logger (*logrus.Logger): Logger for logging rate limiting events.
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler for rate limiting.
func RateLimiterMiddleware(redisClient *redis.Client, limit int, duration time.Duration, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP, exists := c.Get(RealIPKey)
		if !exists {
			logger.Warn("Real IP Middleware not Configured Correctly! Falling Back to RemoteAddr for Rate Limiting!")
			realIP = c.Request.RemoteAddr
		}
		ipAddress := realIP.(string)

		key := "rl:ip:" + ipAddress
		now := time.Now()

		pipe := redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		pipe.PExpireAt(c.Request.Context(), key, now.Add(duration))

		results, err := pipe.Exec(c.Request.Context())
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "ip": ipAddress}).Error("Failed to Increment Request Count in Redis!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		countResult := results[0].(*redis.IntCmd)
		count, err := countResult.Result()
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "ip": ipAddress}).Error("Failed to Get Request Count from Redis!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		if count > int64(limit) {
			retryAfter := duration - time.Since(now)
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			logger.WithFields(logrus.Fields{"ip": ipAddress, "limit": limit, "duration": duration}).Warn("Rate Limit Exceeded!")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests!",
				"message": "Rate Limit Exceeded! Please Try Again after a Minute!",
			})
			return
		}

		c.Next()
	}
}
