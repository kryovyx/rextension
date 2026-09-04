// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares per-route middleware attachment and the priority scale
// that orders the middleware chain.
package rextension

// PerRouteMiddleware decides, for one route, whether this extension's
// middleware applies to it — and if so, with what configuration.
//
// The framework calls the factory **once per route**, when the route table is
// built and every route is known. Returning nil means "not applicable": no
// middleware is attached to that route, ever, so it costs nothing at request
// time rather than being a no-op that still runs.
//
// # Why a factory rather than an event
//
// Extensions previously learned about routes by subscribing to
// router.route.registered, then had to find the route again per request to
// decide whether they applied — which meant keying an index by the raw URL
// path, which meant every parameterized route silently missed. The dependency
// gate never fired for /users/{id}; per-endpoint rate limits fell back to the
// global bucket.
//
// A factory inverts that. The extension declares *how to decide*, and the
// framework supplies each route at the one moment the whole table is known.
// There is nothing to look up at request time and nothing to key by path.
//
// # Capture configuration, do not look it up
//
// Read whatever the middleware needs from rt inside the factory and close over
// it. The closure is built once per route, so per-request work disappears and
// the configuration cannot change under a request that is already running:
//
//	r.UsePerRoute(func(info RouteInfo) Middleware {
//	    gated, ok := info.Route.(DependencyGatedRoute)
//	    if !ok {
//	        return nil // not gated — nothing attached
//	    }
//	    deps := gated.Dependencies() // read ONCE, here
//	    return func(next http.Handler) http.Handler { ... }
//	}, PriorityHealthGate)
//
// # Why RouteInfo rather than route.Route
//
// The factory receives the route *and* the router it is registered on. Some
// middleware needs both: the rate limiter's precedence chain is
// endpoint → router → global, so which limit applies to a route depends on
// which router serves it. With only the route, the extension had to guess —
// it hardcoded the default router's name, so a per-router limit configured
// for any other router was never consulted.
type PerRouteMiddleware func(rt RouteInfo) Middleware

// PerRouterMiddleware decides, for one router, whether this extension's
// middleware applies to it — and if so, with what configuration.
//
// The framework calls the factory once per router, when the route tables are
// built. Returning nil attaches nothing to that router.
//
// This exists for middleware whose configuration depends on the *router* it
// runs on, rather than on the route. The in-flight request gauge is the
// motivating case: it is labelled by router name, so the label has to be
// resolved once per router at composition time. With only Use available, the
// alternatives were both wrong — hardcode one router's name and mislabel every
// other router's requests, or look the name up per request and pay a map
// lookup and a lock for a value that never changes.
type PerRouterMiddleware func(routerName string) Middleware

// The Priority* scale that stood here is in corex now (W22), because WSX's
// handshake chain reuses the same numbers — see middleware.go for the
// re-exports and for why a constant is re-declared rather than aliased.
