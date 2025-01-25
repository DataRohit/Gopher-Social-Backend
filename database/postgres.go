package database

import (
	"fmt"
	"os"
	"strconv"

	"github.com/datarohit/gopher-social-backend/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PostgresDB *gorm.DB

// InitPostgres initializes a connection to PostgreSQL using GORM and stores the DB connection in the PostgresDB variable.
// It also checks if the connection was successful and logs the result.
//
// Parameters:
//   - logger (*logrus.Logger): The logger instance to log the result of the connection.
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
		postgresPort = 5432
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDBName, postgresSSLMode)

	PostgresDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to connect to PostgreSQL with GORM!")
		os.Exit(1)
	}

	db, err := PostgresDB.DB()
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to get underlying SQL DB from GORM!")
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to ping PostgreSQL database!")
		os.Exit(1)
	}

	logger.Info("Connected to PostgreSQL with GORM successfully!")
}

// ClosePostgres closes the connection to PostgreSQL if it is not nil.
// It also logs any errors that occur while closing the connection.
//
// Parameters:
//   - logger (*logrus.Logger): The logger instance to log any errors that occur while closing the connection.
//
// Returns:
//   - None
func ClosePostgres(logger *logrus.Logger) {
	if PostgresDB != nil {
		db, err := PostgresDB.DB()
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Error("Failed to get underlying SQL DB from GORM for closing!")
			return
		}
		if err := db.Close(); err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Error("Error Closing PostgreSQL Connection!")
		} else {
			logger.Info("PostgreSQL connection closed successfully!")
		}
	}
}
