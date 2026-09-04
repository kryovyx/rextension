// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the logger contract from corex, where it moved so that
// one logger can be handed to a router and to a WebSocket gateway in the same
// process and produce one stream (W22).
package rextension

import "github.com/kryovyx/corex"

// LogLevel represents the logging level. Aliased from corex.
type LogLevel = corex.LogLevel

// Log levels, re-declared by value: a constant has no alias form.
const (
	// LogLevelTrace is the most verbose level.
	LogLevelTrace = corex.LogLevelTrace
	// LogLevelDebug is for debug messages.
	LogLevelDebug = corex.LogLevelDebug
	// LogLevelInfo is for informational messages.
	LogLevelInfo = corex.LogLevelInfo
	// LogLevelWarn is for warning messages.
	LogLevelWarn = corex.LogLevelWarn
	// LogLevelError is for error messages.
	LogLevelError = corex.LogLevelError
	// LogLevelOff disables all logging.
	LogLevelOff = corex.LogLevelOff
)

// Logger defines the logging interface for the Rex framework.
// Aliased from corex; the full implementation lives in
// github.com/kryovyx/rex/logger.
type Logger = corex.Logger
