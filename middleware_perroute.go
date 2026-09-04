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

// Middleware chain priorities.
//
// Lower runs further out: PriorityRecovery wraps everything, PriorityDefault is
// innermost, closest to the handler. Middleware with equal priority keeps
// registration order.
//
// The scale is fixed rather than left to registration order because the order
// is *semantic*, not a matter of taste. Getting it wrong produces bugs that do
// not look like ordering bugs:
//
//   - Rate limiting must precede authentication, or a flood of unauthenticated
//     requests each costs a password hash or a token verification before being
//     rejected. The limiter then protects nothing; it just adds work.
//   - CSRF must follow authentication, because a double-submit check needs the
//     session the authenticator established.
//   - Validation must follow both, or bodies get parsed for requests that were
//     always going to be refused.
//   - Recovery must be outermost, or a panic anywhere below it escapes to
//     net/http and kills the connection with no response written.
//
// Leaving that to whichever order New() happened to receive its extensions in
// guarantees a subtle production bug eventually, and one that reproduces only
// on the machine where the arguments were reordered.
const (
	// PriorityRecovery is the outermost position: catch panics from everything
	// inside, including other middleware.
	PriorityRecovery = 100

	// PriorityCORS answers preflights and decorates responses before any
	// authentication or rate limiting rejects them — a browser needs the CORS
	// headers on the error response too, or it reports an opaque failure
	// instead of the actual status.
	PriorityCORS = 200

	// PriorityRateLimit rejects floods before anything expensive happens.
	PriorityRateLimit = 300

	// PriorityAuth authenticates the request and establishes the principal.
	PriorityAuth = 400

	// PriorityCSRF verifies the request came from an allowed origin. After
	// auth: it needs the session.
	PriorityCSRF = 450

	// PriorityValidation parses and validates the request. After auth and rate
	// limiting, so rejected requests are never parsed.
	PriorityValidation = 500

	// PriorityHealthGate refuses requests whose declared dependencies are
	// unavailable. Late, because it is about this route's backing services
	// rather than about the caller.
	PriorityHealthGate = 600

	// PriorityDefault is the innermost position, closest to the handler. Use
	// it for anything that only observes — metrics, tracing, access logging.
	PriorityDefault = 1000
)
