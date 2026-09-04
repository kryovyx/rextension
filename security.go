// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the security scheme contract from corex.
//
// The contract exists so the security extension can publish schemes for the
// OpenAPI extension to document without either importing the other. It moved
// because a WebSocket handshake authenticates with the same schemes over the
// same HTTP request, so wsxtension-asyncapi documents them the same way — and
// a second declaration would mean two vocabularies for one set of credentials.
//
// # The package-level registry is gone (D21)
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
package rextension

import "github.com/kryovyx/corex"

// SecuredRouteAccessor is the minimal interface a route may implement to
// declare which security schemes are required. Aliased from corex.
type SecuredRouteAccessor = corex.SecuredRouteAccessor

// SecuritySchemeAccessor is the minimal interface a security scheme must
// expose so that a document generator can describe it without importing the
// security extension. Aliased from corex.
type SecuritySchemeAccessor = corex.SecuritySchemeAccessor

// ParameterizedScheme is implemented by a security scheme carried in a named
// request parameter — an API key in a header, query string or cookie.
// Aliased from corex.
type ParameterizedScheme = corex.ParameterizedScheme

// BearerFormatProvider is implemented by a bearer scheme that documents the
// token format, e.g. "JWT". Aliased from corex.
type BearerFormatProvider = corex.BearerFormatProvider

// RoleClaimProvider is implemented by a scheme that documents which token
// claim carries the caller's roles. Aliased from corex.
type RoleClaimProvider = corex.RoleClaimProvider

// SchemeRegistry holds the security schemes an application has configured.
// Aliased from corex.
type SchemeRegistry = corex.SchemeRegistry
