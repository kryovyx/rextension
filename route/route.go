// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package route provides routing types for the Rex framework.
//
// This file defines the Route interface and its default implementation.
package route

import (
	"context"
	"net/http"
)

// matchedRouteKey is the unexported context key for the router-matched route.
type matchedRouteKey struct{}

// SetMatchedRoute returns a new context carrying rt as the matched route.
// Called by the router after it resolves a request to a concrete route, so
// middleware running in the same request can access the route without
// re-parsing the URL.
func SetMatchedRoute(ctx context.Context, rt Route) context.Context {
	return context.WithValue(ctx, matchedRouteKey{}, rt)
}

// GetMatchedRoute retrieves the Route stored by SetMatchedRoute.
// Returns the route and true when present; nil and false otherwise.
func GetMatchedRoute(r *http.Request) (Route, bool) {
	rt, ok := r.Context().Value(matchedRouteKey{}).(Route)
	return rt, ok
}

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
