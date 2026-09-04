// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

package event

import (
	"context"
	"net/http"
	"time"
)

// Route is the minimal route interface used in event payloads.
// It is redefined here to keep the event package free of circular imports
// with the root rextension package. Any concrete route type that exposes
// Method and Path satisfies this interface.
type Route interface {
	// Method returns the HTTP method of the route (e.g. "GET", "POST").
	Method() string
	// Path returns the URL path pattern of the route (e.g. "/users/:id").
	Path() string
}

// BaseEvent carries the common fields shared by all events.
type BaseEvent struct {
	eventType string
	ctx       context.Context
	source    string
}

// NewBaseEvent creates a BaseEvent with the given type, source, and context.
// Primarily used to create events for testing.
func NewBaseEvent(ctx context.Context, eventType, source string) BaseEvent {
	return BaseEvent{eventType: eventType, ctx: ctx, source: source}
}

// Type returns the event type identifier.
func (e BaseEvent) Type() string { return e.eventType }

// Context returns the context associated with the event.
func (e BaseEvent) Context() context.Context { return e.ctx }

// Source returns the producer of the event, such as a router name.
func (e BaseEvent) Source() string { return e.source }

// RouterEvent is the base event for router notifications.
type RouterEvent struct {
	BaseEvent
	RouterName string
}

// Name returns the router name associated with the event.
func (e RouterEvent) Name() string { return e.RouterName }

// NewRouterEvent constructs a generic router event for the given router name.
func NewRouterEvent(ctx context.Context, routerName string) RouterEvent {
	return RouterEvent{
		eventType:  RouterEventType,
		ctx:        ctx,
		source:     routerName,
		RouterName: routerName,
	}
}

// RouterInitializedEvent signals that a router has finished initialization.
type RouterInitializedEvent struct {
	RouterEvent
}

// NewRouterInitializedEvent creates a new router initialization event.
func NewRouterInitializedEvent(ctx context.Context, routerName string) RouterInitializedEvent {
	return RouterInitializedEvent{
		eventType: EventTypeRouterInitialized,
		ctx:       ctx,
		// source was not set here, alone among the six constructors, so
		// Source() returned "" for this event and the router name for every
		// other. Three levels of nesting is what hid it; one level is what
		// showed it.
		source:     routerName,
		RouterName: routerName,
	}
}

// RouterRouteRegisteredEvent is emitted when a route is added to a router.
// The Route field is typed as the event.Route interface (Method + Path)
// so that any route implementation satisfies it regardless of handler type.
type RouterRouteRegisteredEvent struct {
	RouterEvent
	Route Route

	// BaseURL is the router's base path prefix, normalised: no trailing
	// slash, and "" for the root (D46).
	//
	// A subscriber that needs to know the URL a route is actually served at
	// has to combine Path with this — the route knows its own pattern but not
	// the prefix its router mounts it under. Without it,
	// rextension-validation carried a getRouterBaseURL helper that returned
	// "" with a comment explaining the event did not carry the value, so its
	// route index was keyed by the wrong path for every non-root router.
	BaseURL string
}

// NewRouterRouteRegisteredEvent constructs a route registration event.
func NewRouterRouteRegisteredEvent(ctx context.Context, routerName string, rt Route, baseURL string) RouterRouteRegisteredEvent {
	return RouterRouteRegisteredEvent{
		eventType:  EventTypeRouterRouteRegistered,
		ctx:        ctx,
		source:     routerName,
		RouterName: routerName,
		Route:      rt,
		BaseURL:    baseURL,
	}
}

// RouterRequestIncomingEvent represents an incoming HTTP request before routing.
type RouterRequestIncomingEvent struct {
	RouterEvent
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

// NewRouterRequestIncomingEvent constructs an incoming-request event instance.
func NewRouterRequestIncomingEvent(ctx context.Context, routerName string, req *http.Request, rw http.ResponseWriter) RouterRequestIncomingEvent {
	return RouterRequestIncomingEvent{
		eventType:      EventTypeRouterRequestIncoming,
		ctx:            ctx,
		source:         routerName,
		RouterName:     routerName,
		Request:        req,
		ResponseWriter: rw,
	}
}

// RouterRequestHandledEvent records metadata about a completed request.
//
// Status, BytesWritten and RoutePattern are new (D17, D27 via O7). Before
// them, a subscriber could observe that a request finished and how long it
// took, but not what it answered — so an access log could not record a status
// code and the metrics extension could not produce
// http_requests_total{status}. The information existed; nothing carried it.
type RouterRequestHandledEvent struct {
	RouterEvent
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Duration       time.Duration

	// Status is the HTTP status code written to the response.
	//
	// Captured by a wrapper the router installs around the ResponseWriter. It
	// is 200 when the handler wrote a body without calling WriteHeader, which
	// is what net/http does.
	Status int

	// BytesWritten is the size of the response body in bytes.
	BytesWritten int64

	// RoutePattern is the matched route's path pattern — "/users/{id}", not
	// "/users/42" — already stripped of the router's BaseURL.
	//
	// This is what a metric label must use. Labelling by Request.URL.Path
	// makes every distinct URL its own time series, which for a parameterized
	// route is unbounded and attacker-controlled: an unauthenticated client
	// requesting /users/1, /users/2, … grows the metric registry without
	// limit (D36).
	//
	// Empty when no route matched.
	RoutePattern string
}

// NewRouterRequestHandledEvent creates an event for a handled request.
func NewRouterRequestHandledEvent(
	ctx context.Context,
	routerName string,
	req *http.Request,
	rw http.ResponseWriter,
	dur time.Duration,
	status int,
	bytesWritten int64,
	routePattern string,
) RouterRequestHandledEvent {
	return RouterRequestHandledEvent{
		eventType:      EventTypeRouterRequestHandled,
		ctx:            ctx,
		source:         routerName,
		RouterName:     routerName,
		Request:        req,
		ResponseWriter: rw,
		Duration:       dur,
		Status:         status,
		BytesWritten:   bytesWritten,
		RoutePattern:   routePattern,
	}
}

// RouterUnresolvedRequestEvent is emitted when no route matches a request.
type RouterUnresolvedRequestEvent struct {
	RouterEvent
	Method string
}

// NewRouterUnresolvedRequestEvent constructs an unresolved-request event.
func NewRouterUnresolvedRequestEvent(ctx context.Context, routerName, method string) RouterUnresolvedRequestEvent {
	return RouterUnresolvedRequestEvent{
		eventType:  EventTypeRouterUnresolvedRequest,
		ctx:        ctx,
		source:     routerName,
		RouterName: routerName,
		Method:     method,
	}
}
