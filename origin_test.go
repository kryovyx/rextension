// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package rextension_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rx "github.com/kryovyx/rextension"
)

func TestOriginPolicy_Allows(t *testing.T) {
	cases := []struct {
		name   string
		policy rx.OriginPolicy
		origin string
		want   bool
	}{
		{"exact match", rx.OriginPolicy{AllowedOrigins: []string{"https://app.example.com"}}, "https://app.example.com", true},
		{"second entry", rx.OriginPolicy{AllowedOrigins: []string{"https://a.example", "https://b.example"}}, "https://b.example", true},
		{"not listed", rx.OriginPolicy{AllowedOrigins: []string{"https://app.example.com"}}, "https://evil.example", false},
		{"empty policy", rx.OriginPolicy{}, "https://app.example.com", false},

		// The suffix/prefix confusions the doc comment calls out. These are the
		// cases exact matching exists to stop, so they are worth pinning.
		{"suffix confusion", rx.OriginPolicy{AllowedOrigins: []string{"https://example.com"}}, "https://evil-example.com", false},
		{"prefix confusion", rx.OriginPolicy{AllowedOrigins: []string{"https://app.example"}}, "https://app.example.attacker.com", false},
		{"subdomain is not implied", rx.OriginPolicy{AllowedOrigins: []string{"https://example.com"}}, "https://sub.example.com", false},
		{"scheme must match", rx.OriginPolicy{AllowedOrigins: []string{"https://app.example.com"}}, "http://app.example.com", false},
		{"port must match", rx.OriginPolicy{AllowedOrigins: []string{"http://localhost:3000"}}, "http://localhost:5173", false},

		// An empty Origin is never allowed — otherwise every allowlist is a
		// no-op for any client that omits the header.
		{"empty origin against wildcard", rx.OriginPolicy{AllowedOrigins: []string{"*"}}, "", false},
		{"empty origin against list", rx.OriginPolicy{AllowedOrigins: []string{"https://a.example"}}, "", false},

		{"wildcard without credentials", rx.OriginPolicy{AllowedOrigins: []string{"*"}}, "https://anything.example", true},
		// Refused rather than emitted: a browser discards the response, so
		// allowing it here would turn a config error into a silent failure.
		{"wildcard with credentials", rx.OriginPolicy{AllowedOrigins: []string{"*"}, AllowCredentials: true}, "https://anything.example", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.policy.Allows(tc.origin); got != tc.want {
				t.Fatalf("Allows(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

// Origin comparison is case-insensitive on scheme and host, which is what
// EqualFold gives — a browser normalises these, but a hand-written allowlist
// may not.
func TestOriginPolicy_Allows_is_case_insensitive(t *testing.T) {
	p := rx.OriginPolicy{AllowedOrigins: []string{"https://App.Example.COM"}}
	if !p.Allows("https://app.example.com") {
		t.Fatal("expected case-insensitive match")
	}
}

// A wildcard anywhere in the list short-circuits, even behind explicit entries.
func TestOriginPolicy_Allows_wildcard_after_explicit_entry(t *testing.T) {
	p := rx.OriginPolicy{AllowedOrigins: []string{"https://a.example", "*"}}
	if !p.Allows("https://unlisted.example") {
		t.Fatal("wildcard later in the list should still allow any origin")
	}
}

func TestOriginPolicy_IsWildcard(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want bool
	}{
		{"bare wildcard", []string{"*"}, true},
		{"wildcard among entries", []string{"https://a.example", "*"}, true},
		{"no wildcard", []string{"https://a.example"}, false},
		{"empty", nil, false},
		{"pattern is not a wildcard", []string{"*.example.com"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := rx.OriginPolicy{AllowedOrigins: tc.in}
			if got := p.IsWildcard(); got != tc.want {
				t.Fatalf("IsWildcard() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOriginPolicy_Valid(t *testing.T) {
	cases := []struct {
		name       string
		policy     rx.OriginPolicy
		wantErr    bool
		wantOrigin string   // the offending entry, when one is at fault
		contains   []string // fragments the message must carry
	}{
		{name: "empty policy is valid", policy: rx.OriginPolicy{}},
		{name: "explicit origins", policy: rx.OriginPolicy{AllowedOrigins: []string{"https://a.example", "http://localhost:3000"}}},
		{name: "wildcard without credentials", policy: rx.OriginPolicy{AllowedOrigins: []string{"*"}}},
		{name: "credentials with explicit origins", policy: rx.OriginPolicy{AllowedOrigins: []string{"https://a.example"}, AllowCredentials: true}},

		{
			name:     "wildcard with credentials",
			policy:   rx.OriginPolicy{AllowedOrigins: []string{"*"}, AllowCredentials: true},
			wantErr:  true,
			contains: []string{"AllowCredentials", "List the origins explicitly"},
		},
		{
			name:       "bare hostname",
			policy:     rx.OriginPolicy{AllowedOrigins: []string{"app.example.com"}},
			wantErr:    true,
			wantOrigin: "app.example.com",
			contains:   []string{"must include a scheme"},
		},
		{
			name:       "trailing path",
			policy:     rx.OriginPolicy{AllowedOrigins: []string{"https://app.example.com/"}},
			wantErr:    true,
			wantOrigin: "https://app.example.com/",
			contains:   []string{"no path, query or fragment"},
		},
		{
			name:       "query string",
			policy:     rx.OriginPolicy{AllowedOrigins: []string{"https://app.example.com?a=1"}},
			wantErr:    true,
			wantOrigin: "https://app.example.com?a=1",
			contains:   []string{"no path, query or fragment"},
		},
		{
			name:       "subdomain pattern",
			policy:     rx.OriginPolicy{AllowedOrigins: []string{"https://*.example.com"}},
			wantErr:    true,
			wantOrigin: "https://*.example.com",
			contains:   []string{"wildcard patterns are not supported"},
		},
		{
			name:       "no host",
			policy:     rx.OriginPolicy{AllowedOrigins: []string{"https://"}},
			wantErr:    true,
			wantOrigin: "https://",
			contains:   []string{"no host"},
		},
		{
			name:     "wildcard entry is skipped by per-entry validation",
			policy:   rx.OriginPolicy{AllowedOrigins: []string{"*", "app.example.com"}},
			wantErr:  true,
			contains: []string{"must include a scheme"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Valid()
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			var pe *rx.OriginPolicyError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *OriginPolicyError, got %T", err)
			}
			if tc.wantOrigin != "" && pe.Origin != tc.wantOrigin {
				t.Fatalf("Origin = %q, want %q", pe.Origin, tc.wantOrigin)
			}
			for _, frag := range tc.contains {
				if !strings.Contains(err.Error(), frag) {
					t.Fatalf("message %q does not contain %q", err.Error(), frag)
				}
			}
		})
	}
}

// The error names the offending entry when there is one, and does not invent an
// entry when the fault is the policy as a whole.
func TestOriginPolicyError_Error(t *testing.T) {
	withOrigin := (&rx.OriginPolicyError{Origin: "app.example.com", Reason: "no scheme"}).Error()
	if !strings.Contains(withOrigin, "app.example.com") || !strings.Contains(withOrigin, "no scheme") {
		t.Fatalf("unexpected message: %q", withOrigin)
	}
	if !strings.HasPrefix(withOrigin, "rextension: invalid allowed origin ") {
		t.Fatalf("unexpected prefix: %q", withOrigin)
	}

	whole := (&rx.OriginPolicyError{Reason: "contradictory"}).Error()
	if !strings.HasPrefix(whole, "rextension: invalid origin policy: ") {
		t.Fatalf("unexpected prefix: %q", whole)
	}
	if strings.Contains(whole, "allowed origin ") {
		t.Fatalf("policy-level error should not name an entry: %q", whole)
	}
}

func TestRequestOrigin(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if got := rx.RequestOrigin(r); got != "" {
		t.Fatalf("expected no origin, got %q", got)
	}
	r.Header.Set("Origin", "https://app.example.com")
	if got := rx.RequestOrigin(r); got != "https://app.example.com" {
		t.Fatalf("got %q", got)
	}
}

// Referer is deliberately not a fallback: it is stripped by privacy settings
// and referrer policies, so honouring it would let client configuration decide
// whether the check applies.
func TestRequestOrigin_ignores_Referer(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Referer", "https://app.example.com/page")
	if got := rx.RequestOrigin(r); got != "" {
		t.Fatalf("Referer must not stand in for Origin, got %q", got)
	}
}

func TestSafeMethod(t *testing.T) {
	safe := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace}
	unsafe := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect}
	for _, m := range safe {
		if !rx.SafeMethod(m) {
			t.Errorf("%s should be safe", m)
		}
	}
	for _, m := range unsafe {
		if rx.SafeMethod(m) {
			t.Errorf("%s should be unsafe", m)
		}
	}
	// Method matching is exact: HTTP methods are case-sensitive, and a
	// lowercase "get" is not a method the router will have matched either.
	if rx.SafeMethod("get") {
		t.Error(`"get" is not GET; SafeMethod must not lowercase-match`)
	}
	if rx.SafeMethod("") {
		t.Error("empty method should not be reported safe")
	}
}
