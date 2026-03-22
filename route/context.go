// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package route defines the Context interface used by route handlers in the Rex framework.
//
// Extensions should import this package for the canonical Context type
// rather than github.com/kryovyx/rex/route.
package route

import (
	"context"
	"net/http"

	"github.com/kryovyx/dix"
)

// Context represents the context of a request or operation within the rex package.
type Context interface {
	context.Context

	// ResponseWriter returns the underlying http.ResponseWriter.
	ResponseWriter() http.ResponseWriter

	// Request returns the inbound HTTP request.
	Request() *http.Request

	// Resolver exposes the DI resolver scoped to this request.
	Resolver() dix.Resolver

	// Respond writes a raw payload with the given status and content type.
	Respond(status int, contentType string, body interface{}) error

	// Text writes a plain text response.
	Text(status int, v string) error

	// JSON writes a JSON-encoded response.
	JSON(status int, v interface{}) error

	// OpenMetrics writes an OpenMetrics-formatted response.
	OpenMetrics(status int, v interface{}) error

	// Param returns a path parameter captured during routing, e.g. "{id}" → Param("id").
	// Returns an empty string when the parameter is not present.
	Param(name string) string

	// SetValue stores a key/value on this context without changing the underlying request context.
	SetValue(key, value interface{})
	// GetValue returns a stored value if present, else falls back to Value().
	GetValue(key interface{}) interface{}
}

// HandlerFunc defines the function signature for route handlers.
type HandlerFunc func(ctx Context)
