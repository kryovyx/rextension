// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Middleware type — the standard Go http middleware signature.
package rextension

import "net/http"

// Middleware is the standard Go HTTP middleware type.
// A Middleware wraps an http.Handler and returns a new http.Handler.
type Middleware func(http.Handler) http.Handler
