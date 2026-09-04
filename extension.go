// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This package provides all the types an extension needs to implement and interact with
// the Rex framework, without depending on the full rex implementation module.
//
// This file declares the Extension interface.
package rextension

import (
	"context"

	"github.com/kryovyx/rextension/route"
)

// Extension represents a Rex framework extension.
// Implement this interface to hook into the Rex application lifecycle.
type Extension interface {
	// OnInitialize is called once when the extension is registered.
	// Use this to set up infrastructure, subscribe to events, and register routes.
	OnInitialize(ctx context.Context, r Rex) error
	// OnStart is called when the Rex application starts.
	OnStart(ctx context.Context, r Rex) error
	// OnReady is called after all listeners have started successfully.
	OnReady(ctx context.Context, r Rex) error
	// OnStop is called when the Rex application is stopping.
	OnStop(ctx context.Context, r Rex) error
	// OnShutdown is called after all resources have been shut down.
	OnShutdown(ctx context.Context, r Rex) error
}

// RouteValidator is implemented by an extension that must check the route
// table before the application serves.
//
// The framework calls ValidateRoutes once, when the table is built — after
// every extension's OnInitialize and OnStart have run, so every route and
// every extension's own configuration are known, and before any listener
// binds. An error aborts startup.
//
// This is the hook that makes a security misconfiguration a deployment
// failure rather than a runtime surprise. Two examples, both previously
// undetectable at startup:
//
//   - A route requiring a security scheme the application never registered.
//     That produced a 500 on every request to the route, discoverable only by
//     making one (D29).
//   - A route declaring roles for a scheme that cannot enforce them. That was
//     served to every authenticated caller, with no diagnostic anywhere — an
//     endpoint marked "admin only" that was not (D28).
//
// Neither is checkable earlier: at OnInitialize the routes do not all exist
// yet, and an extension cannot see another extension's routes at all. Nor
// later: by OnReady the listeners are already accepting.
//
// Implement it on the extension itself:
//
//	func (e *MyExtension) ValidateRoutes(routes []route.Route) error {
//	    for _, rt := range routes { ... }
//	    return nil
//	}
type RouteValidator interface {
	// ValidateRoutes inspects the complete route table. A non-nil error
	// aborts startup and is returned from Run.
	ValidateRoutes(routes []RouteInfo) error
}

// RouteInfo is a route together with the router it is registered on.
//
// The router name is part of the payload because an extension may need it and
// cannot otherwise get it. The OpenAPI generator is the case in point: it
// includes or excludes a route based on which router serves it, so that an
// internal-only listener's routes stay out of the public document.
//
// It used to reconstruct that mapping by subscribing to
// router.route.registered, which is delivered **asynchronously** — the bus
// dispatches on worker goroutines. So the subscription and the synchronous
// startup generation raced, and the generated document could be missing any
// number of routes depending on scheduling. Handing the information over
// directly removes the race rather than papering over it.
type RouteInfo struct {
	// Route is the registered route.
	Route route.Route
	// Router is the name of the router it is registered on.
	Router string
	// BaseURL is that router's normalised base path prefix.
	BaseURL string
}
