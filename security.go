// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the SecuritySchemeAccessor interface and a global registry
// that allows the security extension to publish schemes for use by the OpenAPI
// extension — without either needing to import the other.
package rextension

import "sync"

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

var (
	globalSecuritySchemes []SecuritySchemeAccessor
	globalSchemesMu       sync.Mutex
)

// RegisterSecuritySchemes stores a set of security schemes in a package-level
// registry. Call this from the security extension's OnStart/OnInitialize so
// that the OpenAPI extension can retrieve them at document-generation time
// without any direct import dependency between the two extensions.
func RegisterSecuritySchemes(schemes []SecuritySchemeAccessor) {
	globalSchemesMu.Lock()
	globalSecuritySchemes = make([]SecuritySchemeAccessor, len(schemes))
	copy(globalSecuritySchemes, schemes)
	globalSchemesMu.Unlock()
}

// GetSecuritySchemes returns a snapshot of all registered security schemes.
// Returns nil when no schemes have been registered.
func GetSecuritySchemes() []SecuritySchemeAccessor {
	globalSchemesMu.Lock()
	defer globalSchemesMu.Unlock()
	if len(globalSecuritySchemes) == 0 {
		return nil
	}
	result := make([]SecuritySchemeAccessor, len(globalSecuritySchemes))
	copy(result, globalSecuritySchemes)
	return result
}
