// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Logger interface used by extensions.
package rextension

// LogLevel represents the logging level.
type LogLevel int

const (
	// LogLevelTrace is the most verbose level.
	LogLevelTrace LogLevel = iota
	// LogLevelDebug is for debug messages.
	LogLevelDebug
	// LogLevelInfo is for informational messages.
	LogLevelInfo
	// LogLevelWarn is for warning messages.
	LogLevelWarn
	// LogLevelError is for error messages.
	LogLevelError
	// LogLevelOff disables all logging.
	LogLevelOff
)

// Logger defines the logging interface for the Rex framework.
// The full logger implementation lives in github.com/kryovyx/rex/logger;
// this interface is the canonical source so that extensions can depend on
// rextension only.
type Logger interface {
	// Info logs an informational message.
	Info(format string, args ...interface{})
	// Warn logs a warning message.
	Warn(format string, args ...interface{})
	// Error logs an error message.
	Error(format string, args ...interface{})
	// Debug logs a debug message.
	Debug(format string, args ...interface{})
	// Trace logs a trace message.
	Trace(format string, args ...interface{})
	// SetLogLevel sets the minimum log level.
	SetLogLevel(level LogLevel)
	// WithField returns a logger with an additional field.
	WithField(key string, value interface{}) Logger
	// WithFields returns a logger with additional fields.
	WithFields(fields map[string]interface{}) Logger
	// WithError returns a logger with an error field.
	WithError(err error) Logger
}
