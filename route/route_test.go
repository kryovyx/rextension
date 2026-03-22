// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

package route_test

import (
"testing"

"github.com/kryovyx/rextension/route"
)

// ---- New / defaultRoute ----

func TestNew_ReturnsRoute(t *testing.T) {
handler := route.HandlerFunc(func(route.Context) {})
rt := route.New("GET", "/ping", handler)
if rt == nil {
t.Fatal("expected non-nil Route")
}
}

func TestRoute_Method(t *testing.T) {
rt := route.New("POST", "/users", nil)
if rt.Method() != "POST" {
t.Errorf("Method=%q", rt.Method())
}
}

func TestRoute_Path(t *testing.T) {
rt := route.New("DELETE", "/users/1", nil)
if rt.Path() != "/users/1" {
t.Errorf("Path=%q", rt.Path())
}
}

func TestRoute_Handler_NonNil(t *testing.T) {
called := false
h := route.HandlerFunc(func(route.Context) { called = true })
rt := route.New("GET", "/", h)
if rt.Handler() == nil {
t.Error("Handler should not be nil")
}
rt.Handler()(nil)
if !called {
t.Error("Handler was not invoked")
}
}

func TestRoute_Handler_Nil(t *testing.T) {
rt := route.New("GET", "/health", nil)
if rt.Handler() != nil {
t.Error("expected nil handler")
}
}

func TestNew_AllHTTPMethods(t *testing.T) {
methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
for _, m := range methods {
t.Run(m, func(t *testing.T) {
rt := route.New(m, "/test", nil)
if rt.Method() != m {
t.Errorf("Method=%q, want %q", rt.Method(), m)
}
})
}
}

func TestRoute_ImplementsInterface(t *testing.T) {
var _ route.Route = route.New("GET", "/", nil)
}
