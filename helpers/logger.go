package helpers

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new logger and returns it.
//
// Returns:
//   - logger (*logrus.Logger): The logger instance.
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	formatter := &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logger.SetFormatter(formatter)
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
