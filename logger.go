// logger.go

package pstlib

import (
	"io"
	"log"
	"os"
)

// Logger wraps the standard log.Logger and provides logging levels.
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	// Open the log file for writing, create it if it doesn't exist, append to it if it does.
	logFile, err := os.OpenFile("logs", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	// Create separate loggers for info, error, and debug levels.
	// Info and Debug messages will go to stdout, Error messages will go to stderr.
	return &Logger{
		infoLogger:  log.New(io.MultiWriter(os.Stdout, logFile), "INFO: ", log.LstdFlags),
		errorLogger: log.New(io.MultiWriter(os.Stderr, logFile), "ERROR: ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
	}
}

// Info logs informational messages.
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Error logs error messages.
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Debug logs debug messages.
func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}