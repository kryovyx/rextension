// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the middleware type and the priority scale from corex.
//
// They moved because a WebSocket handshake is an ordinary HTTP GET, so a
// gateway's handshake chain is composed of the same middleware a router's is,
// sorted by the same numbers (W13). Two declarations of one signature would
// let a shared middleware sort differently in the two frameworks, which is a
// bug with no compile error anywhere.
package rextension

import "github.com/kryovyx/corex"

// Middleware is the standard Go HTTP middleware type.
// A Middleware wraps an http.Handler and returns a new http.Handler.
//
// Aliased from corex, so rextension.Middleware and corex.Middleware are one
// type and a middleware written against either satisfies both.
type Middleware = corex.Middleware

// Middleware chain priorities.
//
// Lower runs further out: PriorityRecovery wraps everything, PriorityDefault is
// innermost, closest to the handler. Middleware with equal priority keeps
// registration order. See corex for why the scale is fixed rather than left to
// registration order, and what each ordering constraint protects.
//
// Re-declared by value rather than aliased, because a constant has no alias
// form. The values are therefore duplicated here, and corex's own test asserts
// the ordering they encode — a divergence would be a silent reordering of the
// chain, so it is worth the reader knowing there are two copies of the numbers
// and one authority for their order.
const (
	// PriorityRecovery is the outermost position: catch panics from everything
	// inside, including other middleware.
	PriorityRecovery = corex.PriorityRecovery

	// PriorityCORS answers preflights and decorates responses before any
	// authentication or rate limiting rejects them.
	PriorityCORS = corex.PriorityCORS

	// PriorityRateLimit rejects floods before anything expensive happens.
	PriorityRateLimit = corex.PriorityRateLimit

	// PriorityAuth authenticates the request and establishes the principal.
	PriorityAuth = corex.PriorityAuth

	// PriorityCSRF verifies the request came from an allowed origin. After
	// auth: it needs the session.
	PriorityCSRF = corex.PriorityCSRF

	// PriorityValidation parses and validates the request. After auth and rate
	// limiting, so rejected requests are never parsed.
	PriorityValidation = corex.PriorityValidation

	// PriorityHealthGate refuses requests whose declared dependencies are
	// unavailable.
	PriorityHealthGate = corex.PriorityHealthGate

	// PriorityDefault is the innermost position, closest to the handler.
	PriorityDefault = corex.PriorityDefault
)
