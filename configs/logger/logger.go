package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Level represents log level
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Logger is a simple structured logger
type Logger struct {
	service string
}

// Entry represents a log entry
type Entry struct {
	Time    string                 `json:"time"`
	Level   Level                  `json:"level"`
	Service string                 `json:"service"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

var defaultLogger *Logger

func init() {
	defaultLogger = New("app")
}

// New creates a new logger with service name
func New(service string) *Logger {
	return &Logger{service: service}
}

// Worker returns a logger for worker
func Worker() *Logger {
	return New("worker")
}

// Payment returns a logger for payment notifications
func Payment() *Logger {
	return New("payment-notification")
}

func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	entry := Entry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   level,
		Service: l.service,
		Message: message,
		Fields:  fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("%s [%s] %s\n", entry.Time, entry.Level, entry.Message)
		return
	}

	fmt.Println(string(data))
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(LevelInfo, message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(LevelError, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(LevelWarn, message, fields)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	if os.Getenv("APP_ENV") == "development" {
		l.log(LevelDebug, message, fields)
	}
}

// Info logs an info message using default logger
func Info(message string, fields map[string]interface{}) {
	defaultLogger.Info(message, fields)
}

// Error logs an error message using default logger
func Error(message string, err error, fields map[string]interface{}) {
	defaultLogger.Error(message, err, fields)
}

// Warn logs a warning message using default logger
func Warn(message string, fields map[string]interface{}) {
	defaultLogger.Warn(message, fields)
}

// Debug logs a debug message using default logger
func Debug(message string, fields map[string]interface{}) {
	defaultLogger.Debug(message, fields)
}
