package server

import (
	"bytes"
	"fmt"
	"time"
)

// LogWriter is an io.Writer that sends log messages to a LogBroadcaster.
type LogWriter struct {
	broadcaster *LogBroadcaster
	component   string
}

// NewLogWriter creates a new LogWriter.
func NewLogWriter(lb *LogBroadcaster, component string) *LogWriter {
	return &LogWriter{
		broadcaster: lb,
		component:   component,
	}
}

// Write implements the io.Writer interface.
func (lw *LogWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	p = bytes.TrimSuffix(p, []byte("\n"))

	// For now, assume all messages from the standard log package are INFO level.
	// This will be refined later when we introduce a custom logger with levels.
	msg := LogMessage{
		Type:      "log_entry",
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: lw.component,
		Message:   string(p),
	}
	lw.broadcaster.Broadcast(msg)
	return len(p), nil
}

// CustomLogger wraps the standard log.Logger to add level and component context.
type CustomLogger struct {
	component   string
	broadcaster *LogBroadcaster
}

// NewCustomLogger creates a new CustomLogger.
func NewCustomLogger(lb *LogBroadcaster, component string) *CustomLogger {
	return &CustomLogger{
		component:   component,
		broadcaster: lb,
	}
}

// Info logs an informational message.
func (cl *CustomLogger) Info(format string, v ...interface{}) {
	cl.broadcaster.Broadcast(LogMessage{
		Type:      "log_entry",
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: cl.component,
		Message:   formatMessage(format, v...),
	})
}

// Warn logs a warning message.
func (cl *CustomLogger) Warn(format string, v ...interface{}) {
	cl.broadcaster.Broadcast(LogMessage{
		Type:      "log_entry",
		Timestamp: time.Now(),
		Level:     "WARN",
		Component: cl.component,
		Message:   formatMessage(format, v...),
	})
}

// Error logs an error message.
func (cl *CustomLogger) Error(format string, v ...interface{}) {
	cl.broadcaster.Broadcast(LogMessage{
		Type:      "log_entry",
		Timestamp: time.Now(),
		Level:     "ERROR",
		Component: cl.component,
		Message:   formatMessage(format, v...),
	})
}

// Debug logs a debug message.
func (cl *CustomLogger) Debug(format string, v ...interface{}) {
	cl.broadcaster.Broadcast(LogMessage{
		Type:      "log_entry",
		Timestamp: time.Now(),
		Level:     "DEBUG",
		Component: cl.component,
		Message:   formatMessage(format, v...),
	})
}

// formatMessage formats a message string with optional arguments.
func formatMessage(format string, v ...interface{}) string {
	if len(v) == 0 {
		return format
	}
	return fmt.Sprintf(format, v...)
}
