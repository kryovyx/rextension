// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the sentinel errors the framework returns to extensions.
package rextension

import "errors"

// Sentinel errors returned by the framework to extension code.
//
// They live here, in the contract module, rather than in the framework
// implementation, because an extension depends on `rextension` and not on
// `rex` — so a sentinel declared only in `rex` is unreachable from the code
// that needs to branch on it.
//
// Without them, extensions had to match on message text:
//
//	if !strings.Contains(err.Error(), "already exists") { ... }
//
// which appeared in both rextension-health and rextension-metric, and coupled
// each of them to the exact wording of another module's error string (D39).
var (
	// ErrRouterExists means CreateRouter was called for a name already taken.
	//
	// Usually not a failure: two extensions may both want a router, and
	// whichever runs second should reuse it rather than abort.
	//
	//	if err := r.CreateRouter(name, cfg); err != nil &&
	//	    !errors.Is(err, rextension.ErrRouterExists) {
	//	    return err
	//	}
	ErrRouterExists = errors.New("rex: router already exists")

	// ErrRouterUnknown means a route or middleware named a router that was
	// never created.
	//
	// Reported when the route table is built rather than at the registration
	// call, because whether a router exists depends on every extension having
	// had its turn to create one.
	ErrRouterUnknown = errors.New("rex: unknown router")

	// ErrRouterFrozen means a route or middleware was registered after the
	// route table was built.
	//
	// The table is published by atomic swap and read without a lock, so there
	// is nowhere safe to add to it. Register from OnInitialize or OnStart,
	// never from OnReady.
	ErrRouterFrozen = errors.New("rex: router is already serving; routes and middleware must be registered before Run")
)
