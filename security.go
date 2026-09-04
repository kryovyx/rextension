// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the SecuritySchemeAccessor interface and a global registry
// that allows the security extension to publish schemes for use by the OpenAPI
// extension — without either needing to import the other.
package rextension

// SecuredRouteAccessor is the minimal interface a route may implement to declare
// which security schemes are required. Mirrored here so OpenAPI and Security
// extensions share the type without importing each other.
type SecuredRouteAccessor interface {
	// RequiredSchemes returns the names of the schemes that must authenticate.
	// An empty or nil slice means the route is public.
	RequiredSchemes() []string
}

// SecuritySchemeAccessor is the minimal interface a security scheme must expose
// so that the OpenAPI extension can document it without importing the security
// extension directly.
type SecuritySchemeAccessor interface {
	// Name returns the unique identifier for the scheme (e.g. "bearer", "basic").
	Name() string
	// Type returns the OpenAPI security scheme type (e.g. "http", "apiKey").
	Type() string
	// Description returns a human-readable description for the scheme.
	Description() string
	// Challenge returns the WWW-Authenticate challenge value (e.g. "Bearer").
	Challenge() string
}

// The package-level scheme registry that used to live here is gone (D21).
//
// It was a package-level slice written by RegisterSecuritySchemes and read by
// GetSecuritySchemes. Three problems, in increasing order of severity:
//
//  1. RegisterSecuritySchemes **replaced** the slice rather than appending, and
//     there was no unregister. Two Rex instances in one process therefore
//     clobbered each other's schemes — whichever started last won, for both.
//  2. The state outlived any single application, so it leaked between tests in
//     the same binary: a test that registered schemes changed the result of
//     every later test that read them.
//  3. Nothing owned it, so nothing could reset it.
//
// The replacement is an instance: the security extension registers a
// SchemeRegistry in the DI container, and the OpenAPI extension resolves it
// through the SchemeRegistry interface. Same decoupling — neither extension
// imports the other — with a lifetime bounded by the application that created
// it.
