package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LoggerWriter io.Writer

// SetupLogger mengatur logger dengan rotating log files
func SetupLogger() {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}

	logFile := filepath.Join(logDir, "app.log")

	// Setup rotating log file
	LoggerWriter = &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	// Write to both file and console
	LoggerWriter = io.MultiWriter(LoggerWriter, os.Stdout)

	log.SetOutput(LoggerWriter)
	log.SetLevel(log.LevelInfo)
}

