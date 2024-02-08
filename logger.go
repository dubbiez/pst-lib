// logger.go

package main

import (
	"log"
	"os"
)

// Logger wraps the standard log.Logger and provides logging levels.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs informational messages.
func (l *Logger) Info(v ...interface{}) {
	l.SetPrefix("INFO: ")
	l.Println(v...)
}

// Error logs error messages.
func (l *Logger) Error(v ...interface{}) {
	l.SetPrefix("ERROR: ")
	l.Println(v...)
}

// Debug logs debug messages.
func (l *Logger) Debug(v ...interface{}) {
	l.SetPrefix("DEBUG: ")
	l.Println(v...)
}