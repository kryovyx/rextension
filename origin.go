// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the origin policy from corex, where it moved to gain a
// third consumer: WSX requires one (W12), because WebSocket is not subject to
// CORS — the browser sends Origin and sends the request regardless, with no
// preflight and no opt-in. One allowlist, now three consumers, still written
// once (D33/O11).
package rextension

import (
	"net/http"

	"github.com/kryovyx/corex"
)

// OriginPolicy is an allowlist of browser origins. Aliased from corex.
type OriginPolicy = corex.OriginPolicy

// OriginPolicyError describes an invalid origin policy. Aliased from corex.
//
// Its message now reads "corex: …" rather than "rextension: …", because corex
// is the module that declares it. Nothing branches on the text —
// rextension-cors matches the type with errors.As — so this is the whole of
// the observable change.
type OriginPolicyError = corex.OriginPolicyError

// RequestOrigin returns the request's Origin header.
//
// A helper rather than a bare header read, so every consumer agrees on what
// counts as "no origin". The Referer header is deliberately not consulted as a
// fallback; see corex for why.
func RequestOrigin(r *http.Request) string {
	return corex.RequestOrigin(r)
}

// SafeMethod reports whether a method is safe in the RFC 9110 sense: it does
// not change server state.
func SafeMethod(method string) bool {
	return corex.SafeMethod(method)
}
