// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package route provides routing types for the Rex framework.
//
// This file defines the Route interface and its default implementation.
package route

// Route represents a single route in the application.
type Route interface {
	// Method returns the HTTP method for the route (e.g., GET, POST).
	Method() string
	// Path returns the URL path for the route.
	Path() string
	// Handler returns the handler function for the route.
	Handler() HandlerFunc
}

// defaultRoute is a default implementation of the Route interface.
type defaultRoute struct {
	method  string
	path    string
	handler HandlerFunc
}

// Method returns the HTTP method for the route.
func (r *defaultRoute) Method() string {
	return r.method
}

// Path returns the URL path for the route.
func (r *defaultRoute) Path() string {
	return r.path
}

// Handler returns the handler function for the route.
func (r *defaultRoute) Handler() HandlerFunc {
	return r.handler
}

// New creates a new route with the given method, path, and handler.
func New(method, path string, handler HandlerFunc) Route {
	return &defaultRoute{
		method:  method,
		path:    path,
		handler: handler,
	}
}
