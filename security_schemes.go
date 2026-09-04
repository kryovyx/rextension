// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the optional capabilities a security scheme may expose,
// and the registry interface that replaces the package-level one.
package rextension

// ParameterizedScheme is implemented by a security scheme carried in a named
// request parameter — an API key in a header, query string or cookie.
//
// Location returns a plain string. It used to return a named string type
// (security.APIKeyLocation), which is why the OpenAPI generator reached it with
// reflect.MethodByName("Location") and formatted the result with %s: a named
// type cannot satisfy an interface declaring `Location() string`. Returning
// string makes the interface expressible and deletes the reflection (D20).
type ParameterizedScheme interface {
	// ParamName returns the name of the header, query parameter or cookie.
	ParamName() string
	// Location returns where the parameter is carried: "header", "query" or
	// "cookie".
	Location() string
}

// BearerFormatProvider is implemented by a bearer scheme that documents the
// token format, e.g. "JWT".
type BearerFormatProvider interface {
	// BearerFormat returns the OpenAPI bearerFormat value.
	BearerFormat() string
}

// RoleClaimProvider is implemented by a scheme that documents which token
// claim carries the caller's roles.
type RoleClaimProvider interface {
	// RolesClaim returns the claim name, or "" when the scheme does not
	// expose roles.
	RolesClaim() string
}

// SchemeRegistry holds the security schemes an application has configured.
//
// It replaces a package-level slice with an instance registered in the
// container. The global was written by RegisterSecuritySchemes, which
// *replaced* rather than appended and had no unregister — so two Rex instances
// in one process clobbered each other's schemes, and state leaked between
// tests in the same binary (D21).
type SchemeRegistry interface {
	// Register adds schemes to the registry, ignoring nil entries and
	// duplicate names.
	Register(schemes ...SecuritySchemeAccessor)

	// Schemes returns a snapshot of the registered schemes, ordered by
	// registration.
	Schemes() []SecuritySchemeAccessor

	// Lookup returns the scheme registered under name.
	Lookup(name string) (SecuritySchemeAccessor, bool)
}
