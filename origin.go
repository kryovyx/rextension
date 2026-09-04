// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the origin policy shared by the CORS and CSRF
// implementations (D33/O11).
//
// # Why one policy, two consumers
//
// CORS and CSRF both need to know which origins an application trusts, and
// that list should be written once. Beyond the list they have nothing in
// common, and it is worth being explicit about why — because the belief that
// CORS prevents CSRF is widespread and wrong.
//
// **CORS does not prevent CSRF.** A "simple" request — a form POST with
// application/x-www-form-urlencoded, multipart/form-data or text/plain — gets
// **no preflight**. The browser dispatches it, the server receives it, and the
// server commits the state change. CORS then prevents the attacker from
// *reading* the response, which is no consolation for a transfer that already
// happened.
//
// So CORS answers "may this origin read my responses?" and CSRF answers "did
// this request really come from my own application?". Different questions, one
// allowlist.
//
// The declaration lives here so neither module imports the other.
package rextension

import (
	"net/http"
	"strings"
)

// OriginPolicy is an allowlist of browser origins.
type OriginPolicy struct {
	// AllowedOrigins are matched exactly, scheme and port included:
	// "https://app.example.com", not "app.example.com".
	//
	// Exact matching rather than pattern matching is deliberate. Origin
	// patterns are a classic source of over-permissive policies: a naive
	// suffix check for ".example.com" matches "evil-example.com", and a
	// prefix check for "https://app.example" matches
	// "https://app.example.attacker.com". Both have appeared in real
	// advisories.
	//
	// "*" allows any origin, and is **only** valid without credentials — see
	// AllowCredentials.
	AllowedOrigins []string

	// AllowCredentials permits cookies and HTTP authentication on
	// cross-origin requests.
	//
	// ⚠ Incompatible with a "*" allowlist, and not merely by convention: the
	// Fetch specification requires that Access-Control-Allow-Origin echo a
	// concrete origin when Access-Control-Allow-Credentials is true, and
	// browsers reject the response otherwise. `Allows` therefore refuses "*"
	// when this is set, rather than emitting a combination the browser will
	// discard — a silent failure that looks like a server bug from the
	// client's side.
	AllowCredentials bool
}

// Allows reports whether origin is permitted.
//
// An empty origin is never allowed: a same-origin request carries no Origin
// header for most methods, and such a request needs no CORS decision at all.
// Treating "" as allowed would make every allowlist meaningless.
func (p OriginPolicy) Allows(origin string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range p.AllowedOrigins {
		if allowed == "*" {
			// A wildcard with credentials is not a policy a browser will
			// honour, so it is refused here rather than emitted and dropped.
			return !p.AllowCredentials
		}
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

// IsWildcard reports whether the policy allows any origin.
func (p OriginPolicy) IsWildcard() bool {
	for _, allowed := range p.AllowedOrigins {
		if allowed == "*" {
			return true
		}
	}
	return false
}

// Valid reports whether the policy is internally consistent, and why not.
//
// Checked at startup so a contradictory policy stops a deployment rather than
// producing responses browsers silently discard.
func (p OriginPolicy) Valid() error {
	if p.IsWildcard() && p.AllowCredentials {
		return &OriginPolicyError{
			Reason: "AllowedOrigins contains \"*\" while AllowCredentials is true. " +
				"Browsers reject Access-Control-Allow-Origin: * on a credentialed response, " +
				"so this combination silently blocks every cross-origin request it appears to permit. " +
				"List the origins explicitly.",
		}
	}
	for _, origin := range p.AllowedOrigins {
		if origin == "*" {
			continue
		}
		if err := validateOrigin(origin); err != nil {
			return err
		}
	}
	return nil
}

// OriginPolicyError describes an invalid origin policy.
type OriginPolicyError struct {
	// Origin is the offending entry, when one entry is at fault.
	Origin string
	// Reason explains the problem.
	Reason string
}

func (e *OriginPolicyError) Error() string {
	if e.Origin != "" {
		return "rextension: invalid allowed origin " + e.Origin + ": " + e.Reason
	}
	return "rextension: invalid origin policy: " + e.Reason
}

// validateOrigin checks that an entry is a serialised origin rather than a
// URL, a hostname, or a pattern.
//
// Worth checking, because each mistake fails in a way that is hard to notice:
//
//   - "app.example.com" (no scheme) never matches an Origin header, which
//     always carries one. The policy looks configured and allows nothing.
//   - "https://app.example.com/" (trailing path) never matches either, for the
//     same reason.
//   - "*.example.com" is not a pattern this policy supports, and would be read
//     as a literal origin that no browser sends — so it allows nothing while
//     looking like it allows a subdomain tree.
func validateOrigin(origin string) error {
	if !strings.Contains(origin, "://") {
		return &OriginPolicyError{
			Origin: origin,
			Reason: "an origin must include a scheme, e.g. \"https://app.example.com\". " +
				"A bare hostname never matches an Origin header, so the policy would allow nothing",
		}
	}
	if strings.Contains(origin, "*") {
		return &OriginPolicyError{
			Origin: origin,
			Reason: "wildcard patterns are not supported; list each origin exactly. " +
				"Pattern matching on origins is a recurring source of over-permissive policies",
		}
	}
	rest := origin[strings.Index(origin, "://")+3:]
	if strings.ContainsAny(rest, "/?#") {
		return &OriginPolicyError{
			Origin: origin,
			Reason: "an origin is scheme, host and optional port only — no path, query or fragment. " +
				"An Origin header never carries them, so this would match nothing",
		}
	}
	if rest == "" {
		return &OriginPolicyError{Origin: origin, Reason: "no host"}
	}
	return nil
}

// RequestOrigin returns the request's Origin header.
//
// A helper rather than a bare header read, so both consumers agree on what
// counts as "no origin" — and so the reasoning below lives in one place.
//
// The Referer header is deliberately not consulted as a fallback. It is
// stripped by privacy settings, by referrer policies and by some proxies, so
// treating its absence as "no cross-origin request" would be a bypass, and
// treating its presence as authoritative would let a referrer policy weaken
// the check.
func RequestOrigin(r *http.Request) string {
	return r.Header.Get("Origin")
}

// SafeMethod reports whether a method is safe in the RFC 9110 sense: it does
// not change server state.
//
// CSRF protection applies only to unsafe methods. That is not a convenience —
// a GET that changes state is a bug in its own right, and one that CSRF
// protection cannot fix, because a browser will issue it from an <img> tag
// with no way for the server to distinguish it.
func SafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
