// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

package event_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kryovyx/rextension/event"
)

// ---- minimal test implementations ----

type tEvent struct {
	typ string
	ctx context.Context
}

func (e *tEvent) Type() string             { return e.typ }
func (e *tEvent) Context() context.Context { return e.ctx }

type tRoute struct{ m, p string }

func (r *tRoute) Method() string { return r.m }
func (r *tRoute) Path() string   { return r.p }

// ---- constants ----

func TestConstants(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"RouterEventType", event.RouterEventType, "router_event"},
		{"EventTypeRouterInitialized", event.EventTypeRouterInitialized, "router.initialized"},
		{"EventTypeRouterRouteRegistered", event.EventTypeRouterRouteRegistered, "router.route.registered"},
		{"EventTypeRouterRequestIncoming", event.EventTypeRouterRequestIncoming, "router.request.incoming"},
		{"EventTypeRouterRequestHandled", event.EventTypeRouterRequestHandled, "router.request.handled"},
		{"EventTypeRouterUnresolvedRequest", event.EventTypeRouterUnresolvedRequest, "router.request.unresolved"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val != tt.want {
				t.Errorf("got %q, want %q", tt.val, tt.want)
			}
		})
	}
}

// ---- As ----

func TestAs_Success(t *testing.T) {
	ctx := context.Background()
	ev := &tEvent{typ: "foo", ctx: ctx}
	var e event.Event = ev
	got, ok := event.As[*tEvent](e)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got != ev {
		t.Error("value mismatch")
	}
}

func TestAs_Failure(t *testing.T) {
	ev := &tEvent{typ: "foo", ctx: context.Background()}
	var e event.Event = ev
	_, ok := event.As[event.BaseEvent](e)
	if ok {
		t.Error("expected ok=false")
	}
}

// ---- BaseEvent ----

func TestNewBaseEvent(t *testing.T) {
	ctx := context.Background()
	be := event.NewBaseEvent(ctx, "my.type", "src1")
	if be.Type() != "my.type" {
		t.Errorf("Type=%q", be.Type())
	}
	if be.Context() != ctx {
		t.Error("Context mismatch")
	}
	if be.Source() != "src1" {
		t.Errorf("Source=%q", be.Source())
	}
}

func TestBaseEvent_NilContext(t *testing.T) {
	be := event.NewBaseEvent(nil, "t", "s")
	if be.Context() != nil {
		t.Error("expected nil context")
	}
}

// ---- RouterEvent ----

func TestNewRouterEvent(t *testing.T) {
	ctx := context.Background()
	re := event.NewRouterEvent(ctx, "main")
	if re.Type() != event.RouterEventType {
		t.Errorf("Type=%q", re.Type())
	}
	if re.Context() != ctx {
		t.Error("Context mismatch")
	}
	if re.Source() != "main" {
		t.Errorf("Source=%q", re.Source())
	}
	if re.Name() != "main" {
		t.Errorf("Name=%q", re.Name())
	}
}

// ---- RouterInitializedEvent ----

func TestNewRouterInitializedEvent(t *testing.T) {
	ctx := context.Background()
	e := event.NewRouterInitializedEvent(ctx, "api")
	if e.Type() != event.EventTypeRouterInitialized {
		t.Errorf("Type=%q", e.Type())
	}
	if e.Context() != ctx {
		t.Error("Context mismatch")
	}
	if e.RouterName != "api" {
		t.Errorf("RouterName=%q", e.RouterName)
	}
}

// ---- RouterRouteRegisteredEvent ----

func TestNewRouterRouteRegisteredEvent(t *testing.T) {
	ctx := context.Background()
	rt := &tRoute{m: "POST", p: "/items"}
	e := event.NewRouterRouteRegisteredEvent(ctx, "default", rt, "")
	if e.Type() != event.EventTypeRouterRouteRegistered {
		t.Errorf("Type=%q", e.Type())
	}
	if e.Context() != ctx {
		t.Error("Context mismatch")
	}
	if e.RouterName != "default" {
		t.Errorf("RouterName=%q", e.RouterName)
	}
	if e.Route.Method() != "POST" || e.Route.Path() != "/items" {
		t.Error("Route mismatch")
	}
}

// ---- RouterRequestIncomingEvent ----

func TestNewRouterRequestIncomingEvent(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rw := httptest.NewRecorder()
	e := event.NewRouterRequestIncomingEvent(ctx, "router1", req, rw)
	if e.Type() != event.EventTypeRouterRequestIncoming {
		t.Errorf("Type=%q", e.Type())
	}
	if e.Context() != ctx {
		t.Error("Context mismatch")
	}
	if e.RouterName != "router1" {
		t.Errorf("RouterName=%q", e.RouterName)
	}
	if e.Source() != "router1" {
		t.Errorf("Source=%q", e.Source())
	}
	if e.Request != req {
		t.Error("Request mismatch")
	}
	if e.ResponseWriter != rw {
		t.Error("ResponseWriter mismatch")
	}
}

// ---- RouterRequestHandledEvent ----

func TestNewRouterRequestHandledEvent(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	rw := httptest.NewRecorder()
	dur := 42 * time.Millisecond
	e := event.NewRouterRequestHandledEvent(ctx, "main", req, rw, dur, http.StatusOK, 0, "")
	if e.Type() != event.EventTypeRouterRequestHandled {
		t.Errorf("Type=%q", e.Type())
	}
	if e.Context() != ctx {
		t.Error("Context mismatch")
	}
	if e.RouterName != "main" {
		t.Errorf("RouterName=%q", e.RouterName)
	}
	if e.Source() != "main" {
		t.Errorf("Source=%q", e.Source())
	}
	if e.Request != req {
		t.Error("Request mismatch")
	}
	if e.ResponseWriter != rw {
		t.Error("ResponseWriter mismatch")
	}
	if e.Duration != dur {
		t.Errorf("Duration=%v", e.Duration)
	}
}

// ---- RouterUnresolvedRequestEvent ----

func TestNewRouterUnresolvedRequestEvent(t *testing.T) {
	ctx := context.Background()
	e := event.NewRouterUnresolvedRequestEvent(ctx, "api", "PUT")
	if e.Type() != event.EventTypeRouterUnresolvedRequest {
		t.Errorf("Type=%q", e.Type())
	}
	if e.Context() != ctx {
		t.Error("Context mismatch")
	}
	if e.RouterName != "api" {
		t.Errorf("RouterName=%q", e.RouterName)
	}
	if e.Source() != "api" {
		t.Errorf("Source=%q", e.Source())
	}
	if e.Method != "PUT" {
		t.Errorf("Method=%q", e.Method)
	}
}

// Every router event reports the router as its source.
//
// event.NewRouterInitializedEvent did not set source, alone among the six, so
// Source() returned "" for it. Nothing in production read Source(), so this
// pins the behaviour rather than fixing a live failure — and pins it for all
// six together, which is what stops one drifting again.
func TestRouterEvents_all_report_the_router_as_source(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()

	cases := map[string]interface{ Source() string }{
		"event.NewRouterEvent":                  event.NewRouterEvent(ctx, "api"),
		"event.NewRouterInitializedEvent":       event.NewRouterInitializedEvent(ctx, "api"),
		"event.NewRouterRouteRegisteredEvent":   event.NewRouterRouteRegisteredEvent(ctx, "api", nil, "/"),
		"event.NewRouterRequestIncomingEvent":   event.NewRouterRequestIncomingEvent(ctx, "api", req, rec),
		"event.NewRouterRequestHandledEvent":    event.NewRouterRequestHandledEvent(ctx, "api", req, rec, time.Second, 200, 10, "/x"),
		"event.NewRouterUnresolvedRequestEvent": event.NewRouterUnresolvedRequestEvent(ctx, "api", http.MethodGet),
	}
	for name, ev := range cases {
		if got := ev.Source(); got != "api" {
			t.Errorf("%s: Source() = %q, want %q", name, got, "api")
		}
	}
}

// And each still carries its own type and router name after the flattening.
func TestRouterEvents_keep_their_type_and_name(t *testing.T) {
	ctx := context.Background()
	cases := map[string]struct {
		ev       interface{ Type() string }
		wantType string
	}{
		"initialized": {event.NewRouterInitializedEvent(ctx, "api"), event.EventTypeRouterInitialized},
		"registered":  {event.NewRouterRouteRegisteredEvent(ctx, "api", nil, "/"), event.EventTypeRouterRouteRegistered},
		"unresolved":  {event.NewRouterUnresolvedRequestEvent(ctx, "api", http.MethodGet), event.EventTypeRouterUnresolvedRequest},
	}
	for name, tc := range cases {
		if got := tc.ev.Type(); got != tc.wantType {
			t.Errorf("%s: Type() = %q, want %q", name, got, tc.wantType)
		}
	}
	if got := event.NewRouterInitializedEvent(ctx, "api").Name(); got != "api" {
		t.Errorf("Name() = %q", got)
	}
}
