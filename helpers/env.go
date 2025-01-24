package helpers

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// GetEnv returns the value of an environment variable or
// a default value if the environment variable is not set.
// The value is trimmed of leading and trailing whitespace.
//
// Parameters:
//   - env (string): The name of the environment variable.
//   - defaultValue (string): The default value to return if the environment variable is not set.
//
// Returns:
//   - environment (string): The value of the environment variable or the default value.
func GetEnv(env, defaultValue string) string {
	environment := strings.TrimSpace(os.Getenv(env))
	if environment == "" {
		return defaultValue
	}
	return environment
}

// GetEnvAsInt returns the value of an environment variable as an integer or
// a default value if the environment variable is not set or is not a valid integer.
// The value is trimmed of leading and trailing whitespace.
//
// Parameters:
//   - env (string): The name of the environment variable.
//   - defaultValue (int): The default value to return if the environment variable is not set or is not a valid integer.
//
// Returns:
//   - value (int): The value of the environment variable as an integer or the default value.
func GetEnvAsInt(env string, defaultValue int) int {
	environment := strings.TrimSpace(os.Getenv(env))
	if environment == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(environment)
	if err != nil {
		log.Printf("Warning: %s is not a valid integer. Using default value: %d", env, defaultValue)
		return defaultValue
	}

	return value
}
