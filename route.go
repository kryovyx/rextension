// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Route interface used when registering routes via Rex.
package rextension

// Route is the minimal interface for a route that can be registered with the Rex router.
// The concrete route type (e.g. from github.com/kryovyx/rex/route) satisfies this interface.
type Route interface {
	// Method returns the HTTP method for the route (e.g., GET, POST).
	Method() string
	// Path returns the URL path for the route.
	Path() string
}
