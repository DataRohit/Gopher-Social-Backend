package database

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var PostgresDB *pgxpool.Pool

// InitPostgres initializes the PostgreSQL database connection pool.
// It reads database connection parameters from environment variables or uses default values.
//
// Parameters:
//   - logger (*logrus.Logger): Logrus logger instance for logging.
//
// Returns:
//   - None
func InitPostgres(logger *logrus.Logger) {
	postgresHost := helpers.GetEnv("POSTGRES_HOST", "localhost")
	postgresPortStr := helpers.GetEnv("POSTGRES_PORT", "5432")
	postgresUser := helpers.GetEnv("POSTGRES_USER", "postgres")
	postgresPassword := helpers.GetEnv("POSTGRES_PASSWORD", "postgres")
	postgresDBName := helpers.GetEnv("POSTGRES_DBNAME", "gopher")
	postgresSSLMode := helpers.GetEnv("POSTGRES_SSLMODE", "disable")

	postgresPort, err := strconv.Atoi(postgresPortStr)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err, "port": postgresPortStr}).Fatal("Failed to parse PostgreSQL Port!")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDBName, postgresSSLMode)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to parse PostgreSQL connection string!")
		os.Exit(1)
	}

	PostgresDB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to connect to PostgreSQL!")
		os.Exit(1)
	}

	if PostgresDB == nil {
		logger.Fatal("PostgreSQL Database connection pool failed to initialize!")
		os.Exit(1)
	}
	logger.Info("PostgreSQL Database Connected Successfully!")
}

// ClosePostgres closes the PostgreSQL database connection pool.
// It checks if the PostgresDB instance is not nil before attempting to close the connection.
//
// Parameters:
//   - logger (*logrus.Logger): Logrus logger instance for logging.
//
// Returns:
//   - None
func ClosePostgres(logger *logrus.Logger) {
	if PostgresDB != nil {
		PostgresDB.Close()
		logger.Info("PostgreSQL Database Connection Closed Successfully!")
	}
}
