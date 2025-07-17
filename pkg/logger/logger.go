package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// contextKey is a private type to avoid context key collisions
type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

var Global *logrus.Logger

// New sets up the global logrus logger and returns it. It also sets the Global variable.
func New(level string) *logrus.Logger {
	log := logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// Set formatter to JSON for structured logging
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	log.SetOutput(os.Stdout)
	Global = log
	return log
}

// WithLogger returns a new context with the provided logger
func WithLogger(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromContext returns the logger from context, or the global logger if not found
func FromContext(ctx context.Context) *logrus.Entry {
	log, ok := ctx.Value(loggerKey).(*logrus.Entry)
	if !ok || log == nil {
		if Global != nil {
			return logrus.NewEntry(Global)
		}
		return logrus.NewEntry(logrus.New())
	}
	return log
}

// WithRequestID returns a new context with the provided request_id
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext returns the request_id from context, or empty string if not found
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithRequestIDLogger returns a logger.Entry with request_id field if present in context
func WithRequestIDLogger(ctx context.Context) *logrus.Entry {
	log := FromContext(ctx)
	if reqID := RequestIDFromContext(ctx); reqID != "" {
		return log.WithField("request_id", reqID)
	}
	return log
}

// Wrapper functions for convenience logging
func Info(args ...interface{}) {
	if Global != nil {
		Global.Info(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Global != nil {
		Global.Infof(format, args...)
	}
}

func Error(args ...interface{}) {
	if Global != nil {
		Global.Error(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Global != nil {
		Global.Errorf(format, args...)
	}
}

func Debug(args ...interface{}) {
	if Global != nil {
		Global.Debug(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Global != nil {
		Global.Debugf(format, args...)
	}
}

func Warn(args ...interface{}) {
	if Global != nil {
		Global.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Global != nil {
		Global.Warnf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	if Global != nil {
		Global.Fatal(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if Global != nil {
		Global.Fatalf(format, args...)
	}
}
