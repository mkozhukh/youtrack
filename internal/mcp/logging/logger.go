package logging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// KeyHashKey is the context key for the API key hash
	KeyHashKey contextKey = "log_key_hash"
)

// WithKeyHash adds an API key hash to the context
func WithKeyHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, KeyHashKey, hash)
}

// GetKeyHash extracts the API key hash from context
func GetKeyHash(ctx context.Context) string {
	if v := ctx.Value(KeyHashKey); v != nil {
		return v.(string)
	}
	return ""
}

// HashAPIKey creates a short SHA256 hash of an API key for logging
func HashAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])[:12] // First 12 chars of hex
}

// RESTLogger is the interface for logging REST calls
type RESTLogger interface {
	LogRESTCall(method, path string, duration time.Duration)
	LogRESTError(method, path string, body interface{}, statusCode int, errMsg string)
}

// AppLogger provides structured logging to three separate files
type AppLogger struct {
	config LogConfig

	callLogFile      *os.File
	restErrorLogFile *os.File
	toolErrorLogFile *os.File

	callMu      sync.Mutex
	restErrorMu sync.Mutex
	toolErrorMu sync.Mutex
}

// NewAppLogger creates a new AppLogger with the given configuration
func NewAppLogger(config LogConfig) (*AppLogger, error) {
	if !config.Enabled {
		return &AppLogger{config: config}, nil
	}

	logger := &AppLogger{config: config}

	// Open call log file
	if config.CallLogPath != "" {
		f, err := os.OpenFile(config.CallLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Warn("Failed to open call log file, call logging disabled", "path", config.CallLogPath, "error", err)
		} else {
			logger.callLogFile = f
		}
	}

	// Open REST error log file
	if config.RESTErrorLogPath != "" {
		f, err := os.OpenFile(config.RESTErrorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Warn("Failed to open REST error log file, REST error logging disabled", "path", config.RESTErrorLogPath, "error", err)
		} else {
			logger.restErrorLogFile = f
		}
	}

	// Open tool error log file
	if config.ToolErrorLogPath != "" {
		f, err := os.OpenFile(config.ToolErrorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Warn("Failed to open tool error log file, tool error logging disabled", "path", config.ToolErrorLogPath, "error", err)
		} else {
			logger.toolErrorLogFile = f
		}
	}

	return logger, nil
}

// Close closes all log files
func (l *AppLogger) Close() error {
	if l.callLogFile != nil {
		l.callLogFile.Close()
	}
	if l.restErrorLogFile != nil {
		l.restErrorLogFile.Close()
	}
	if l.toolErrorLogFile != nil {
		l.toolErrorLogFile.Close()
	}
	return nil
}

// LogToolCall logs a tool invocation to calls.log
func (l *AppLogger) LogToolCall(keyHash, toolName string) {
	if !l.config.Enabled || l.callLogFile == nil {
		return
	}

	entry := map[string]interface{}{
		"t":    time.Now().UTC().Format(time.RFC3339),
		"type": "tool",
		"key":  keyHash,
		"tool": toolName,
	}

	l.writeCallLog(entry)
}

// LogRESTCall logs a REST API call to calls.log
func (l *AppLogger) LogRESTCall(keyHash, method, path string, duration time.Duration) {
	if !l.config.Enabled || l.callLogFile == nil {
		return
	}

	entry := map[string]interface{}{
		"t":      time.Now().UTC().Format(time.RFC3339),
		"type":   "rest",
		"key":    keyHash,
		"method": method,
		"path":   path,
		"ms":     duration.Milliseconds(),
	}

	l.writeCallLog(entry)
}

// LogRESTError logs a REST API error to rest_errors.log
func (l *AppLogger) LogRESTError(keyHash, method, path string, params interface{}, statusCode int, errMsg string) {
	if !l.config.Enabled || l.restErrorLogFile == nil {
		return
	}

	entry := map[string]interface{}{
		"t":      time.Now().UTC().Format(time.RFC3339),
		"key":    keyHash,
		"method": method,
		"path":   path,
		"params": params,
		"status": statusCode,
		"error":  errMsg,
	}

	l.writeRESTErrorLog(entry)
}

// LogToolError logs a tool error to tool_errors.log
func (l *AppLogger) LogToolError(keyHash, toolName string, params map[string]interface{}, errMsg string) {
	if !l.config.Enabled || l.toolErrorLogFile == nil {
		return
	}

	entry := map[string]interface{}{
		"t":      time.Now().UTC().Format(time.RFC3339),
		"key":    keyHash,
		"tool":   toolName,
		"params": params,
		"error":  errMsg,
	}

	l.writeToolErrorLog(entry)
}

func (l *AppLogger) writeCallLog(entry map[string]interface{}) {
	l.callMu.Lock()
	defer l.callMu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Error("Failed to marshal call log entry", "error", err)
		return
	}
	if _, err := l.callLogFile.WriteString(string(data) + "\n"); err != nil {
		log.Error("Failed to write to call log", "error", err)
	}
}

func (l *AppLogger) writeRESTErrorLog(entry map[string]interface{}) {
	l.restErrorMu.Lock()
	defer l.restErrorMu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Error("Failed to marshal REST error log entry", "error", err)
		return
	}
	if _, err := l.restErrorLogFile.WriteString(string(data) + "\n"); err != nil {
		log.Error("Failed to write to REST error log", "error", err)
	}
}

func (l *AppLogger) writeToolErrorLog(entry map[string]interface{}) {
	l.toolErrorMu.Lock()
	defer l.toolErrorMu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Error("Failed to marshal tool error log entry", "error", err)
		return
	}
	if _, err := l.toolErrorLogFile.WriteString(string(data) + "\n"); err != nil {
		log.Error("Failed to write to tool error log", "error", err)
	}
}

// RESTLoggerWithContext creates a RESTLogger that includes keyHash in all logs
type RESTLoggerWithContext struct {
	logger  *AppLogger
	keyHash string
}

// NewRESTLoggerWithContext creates a REST logger with context
func (l *AppLogger) NewRESTLoggerWithContext(keyHash string) *RESTLoggerWithContext {
	return &RESTLoggerWithContext{
		logger:  l,
		keyHash: keyHash,
	}
}

// LogRESTCall logs a REST call with the keyHash
func (r *RESTLoggerWithContext) LogRESTCall(method, path string, duration time.Duration) {
	if r.logger != nil {
		r.logger.LogRESTCall(r.keyHash, method, path, duration)
	}
}

// LogRESTError logs a REST error with the keyHash
func (r *RESTLoggerWithContext) LogRESTError(method, path string, body interface{}, statusCode int, errMsg string) {
	if r.logger != nil {
		r.logger.LogRESTError(r.keyHash, method, path, body, statusCode, errMsg)
	}
}
