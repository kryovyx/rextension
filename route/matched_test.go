// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package route_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	rxroute "github.com/kryovyx/rextension/route"
)

type stubRoute struct {
	method string
	path   string
}

func (s stubRoute) Method() string               { return s.method }
func (s stubRoute) Path() string                 { return s.path }
func (s stubRoute) Handler() rxroute.HandlerFunc { return func(rxroute.Context) {} }

// The matched route is how every per-route middleware learns which route it is
// running for. Reading it back is the contract; the index that used to stand in
// for it is gone (P4.19).
func TestSetGetMatchedRoute(t *testing.T) {
	rt := stubRoute{method: http.MethodGet, path: "/users/{userId}"}
	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	r = r.WithContext(rxroute.SetMatchedRoute(r.Context(), rt))

	got, ok := rxroute.GetMatchedRoute(r)
	if !ok {
		t.Fatal("expected the matched route to be present")
	}
	// The registered pattern, not the live URL — the distinction that made the
	// old pattern-keyed index always miss.
	if got.Path() != "/users/{userId}" {
		t.Fatalf("Path() = %q, want the registered pattern", got.Path())
	}
	if got.Method() != http.MethodGet {
		t.Fatalf("Method() = %q", got.Method())
	}
}

func TestGetMatchedRoute_absent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rt, ok := rxroute.GetMatchedRoute(r)
	if ok {
		t.Fatal("expected ok=false with no route in the context")
	}
	if rt != nil {
		t.Fatalf("expected a nil route, got %#v", rt)
	}
}

// Later calls replace earlier ones, so a re-dispatch cannot leave a stale route
// visible to middleware.
func TestSetMatchedRoute_overwrites(t *testing.T) {
	ctx := rxroute.SetMatchedRoute(context.Background(), stubRoute{path: "/first"})
	ctx = rxroute.SetMatchedRoute(ctx, stubRoute{path: "/second"})
	r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	got, ok := rxroute.GetMatchedRoute(r)
	if !ok || got.Path() != "/second" {
		t.Fatalf("got %v/%v, want /second", got, ok)
	}
}
