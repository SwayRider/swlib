// Package logger provides context-aware logging with emoji-based log levels.
//
// It supports both package-level functions for simple logging and instance
// methods for component/function-scoped logging with context fields.
//
// # Log Levels
//
//   - Info: General information (ℹ️)
//   - Debug: Debug details (🐛)
//   - Warn: Warnings (⚠️)
//   - Error: Errors (❌)
//   - Fatal: Fatal errors that exit the program (❌)
//   - Success: Success messages (✅)
//
// # Package-Level Logging
//
//	logger.Infof("Application starting on port %d", 8080)
//	logger.Errorf("Failed to connect: %v", err)
//
// # Component-Scoped Logging
//
//	log := logger.New(logger.WithComponent("user-service"))
//	log.Infof("Processing request")
//
//	funcLog := log.Derive(logger.WithFunction("CreateUser"))
//	funcLog.Successf("User created")
package logger

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

// Log level emoji symbols
const (
	SuccessSymbol = "✅"
	ErrorSymbol   = "❌"
	InfoSymbol    = "ℹ️ "
	WarnSymbol    = "⚠️ "
	DebugSymbol   = "🐛"
)

// LogLevel represents the minimum log level to display
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Global log level - controls which messages are shown
var currentLogLevel LogLevel = LogLevelInfo // Default to info

var (
	generalLogger = log.New(os.Stdout, "", log.LstdFlags)
	infoLogger    = log.New(os.Stdout, InfoSymbol+" [INFO] ", log.LstdFlags)
	debugLogger   = log.New(os.Stdout, DebugSymbol+" [DEBUG] ", log.LstdFlags)
	warnLogger    = log.New(os.Stdout, WarnSymbol+" [WARN] ", log.LstdFlags)
	errorLogger   = log.New(os.Stdout, ErrorSymbol+" [ERROR] ", log.LstdFlags)
	successLogger = log.New(os.Stdout, SuccessSymbol+" [SUCCESS] ", log.LstdFlags)
)

// SetLogLevel configures the minimum log level for all loggers.
// Valid values: "debug", "info", "warn", "error"
// Returns error if level string is invalid.
func SetLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		currentLogLevel = LogLevelDebug
	case "info":
		currentLogLevel = LogLevelInfo
	case "warn", "warning":
		currentLogLevel = LogLevelWarn
	case "error":
		currentLogLevel = LogLevelError
	default:
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", level)
	}
	return nil
}

// ContextField represents a key-value pair added to log messages.
// Context fields are prepended to log messages as [key=value].
type ContextField struct {
	Key   string
	Value any
}

// WithComponent creates a context field for component identification.
// Example: logger.New(logger.WithComponent("auth-service"))
func WithComponent(value any) ContextField {
	return WithContextField("component", value)
}

// WithFunction creates a context field for function identification.
// Example: log.Derive(logger.WithFunction("CreateUser"))
func WithFunction(value any) ContextField {
	return WithContextField("function", value)
}

// WithContextField creates a custom context field with the given key and value.
func WithContextField(key string, value any) ContextField {
	return ContextField{
		Key:   key,
		Value: value,
	}
}

func (cf ContextField) Format(msg string) string {
	return fmt.Sprintf("[%s=%v] %s", cf.Key, cf.Value, msg)
}

// Logger provides context-aware logging with attached context fields.
// Context fields are prepended to all log messages from this logger.
type Logger struct {
	contextFields []ContextField
}

// New creates a new Logger with the given context fields.
// Example:
//
//	log := logger.New(logger.WithComponent("auth-service"))
func New(contextFields ...ContextField) *Logger {
	slices.Reverse(contextFields)
	return &Logger{
		contextFields: contextFields,
	}
}

// Derive creates a child logger with additional context fields.
// The new fields are merged with existing fields, with new fields taking precedence.
// Example:
//
//	funcLog := log.Derive(logger.WithFunction("CreateUser"))
func (l *Logger) Derive(contextFields ...ContextField) *Logger {
	slices.Reverse(contextFields)
	return &Logger{
		contextFields: mergeContextFields(l.contextFields, contextFields),
	}
}

func (l Logger) Printf(format string, args ...any) {
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	generalLogger.Printf(format, args...)
}

func (l Logger) Println(msg string) {
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	generalLogger.Println(msg)
}

func (l Logger) Infof(format string, args ...any) {
	if currentLogLevel > LogLevelInfo {
		return
	}
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	infoLogger.Printf(format, args...)
}

func (l Logger) Infoln(msg string) {
	if currentLogLevel > LogLevelInfo {
		return
	}
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	infoLogger.Println(msg)
}

func (l Logger) Warnf(format string, args ...any) {
	if currentLogLevel > LogLevelWarn {
		return
	}
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	warnLogger.Printf(format, args...)
}

func (l Logger) Warnln(msg string) {
	if currentLogLevel > LogLevelWarn {
		return
	}
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	warnLogger.Println(msg)
}

func (l Logger) Debugf(format string, args ...any) {
	if currentLogLevel > LogLevelDebug {
		return
	}
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	debugLogger.Printf(format, args...)
}

func (l Logger) Debugln(msg string) {
	if currentLogLevel > LogLevelDebug {
		return
	}
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	debugLogger.Println(msg)
}

func (l Logger) Errorf(format string, args ...any) {
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	errorLogger.Printf(format, args...)
}

func (l Logger) Errorln(msg string) {
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	errorLogger.Println(msg)
}

func (l Logger) Fatalf(format string, args ...any) {
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	errorLogger.Fatalf(format, args...)
}

func (l Logger) Fatalln(msg string) {
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	errorLogger.Fatalln(msg)
}

func (l Logger) Successf(format string, args ...any) {
	for _, cf := range l.contextFields {
		format = cf.Format(format)
	}
	successLogger.Printf(format, args...)
}

func (l Logger) Successln(msg string) {
	for _, cf := range l.contextFields {
		msg = cf.Format(msg)
	}
	successLogger.Println(msg)
}

func Printf(format string, args ...any) {
	generalLogger.Printf(format, args...)
}

func Println(msg string) {
	generalLogger.Println(msg)
}

func Infof(format string, args ...any) {
	if currentLogLevel > LogLevelInfo {
		return
	}
	infoLogger.Printf(format, args...)
}

func Infoln(msg string) {
	if currentLogLevel > LogLevelInfo {
		return
	}
	infoLogger.Println(msg)
}

func Warnf(format string, args ...any) {
	if currentLogLevel > LogLevelWarn {
		return
	}
	warnLogger.Printf(format, args...)
}

func Warnln(msg string) {
	if currentLogLevel > LogLevelWarn {
		return
	}
	warnLogger.Println(msg)
}

func Debugf(format string, args ...any) {
	if currentLogLevel > LogLevelDebug {
		return
	}
	debugLogger.Printf(format, args...)
}

func Debugln(msg string) {
	if currentLogLevel > LogLevelDebug {
		return
	}
	debugLogger.Println(msg)
}

func Errorf(format string, args ...any) {
	errorLogger.Printf(format, args...)
}

func Errorln(msg string) {
	errorLogger.Println(msg)
}

func Fatalf(format string, args ...any) {
	errorLogger.Fatalf(format, args...)
}

func Fatalln(msg string) {
	errorLogger.Fatalln(msg)
}

func Successf(format string, args ...any) {
	successLogger.Printf(format, args...)
}

func Successln(msg string) {
	successLogger.Println(msg)
}

func mergeContextFields(orig []ContextField, new []ContextField) []ContextField {
	keys := make(map[string]struct{})
	result := make([]ContextField, 0, len(orig)+len(new))
	for _, cf := range new {
		if _, ok := keys[cf.Key]; ok {
			continue
		}
		keys[cf.Key] = struct{}{}
		result = append(result, cf)
	}
	for _, cf := range orig {
		if _, ok := keys[cf.Key]; ok {
			continue
		}
		keys[cf.Key] = struct{}{}
		result = append(result, cf)
	}
	return result
}
