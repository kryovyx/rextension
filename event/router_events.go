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
		BaseEvent: BaseEvent{
			eventType: RouterEventType,
			ctx:       ctx,
			source:    routerName,
		},
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
		RouterEvent: RouterEvent{
			BaseEvent: BaseEvent{
				eventType: EventTypeRouterInitialized,
				ctx:       ctx,
			},
			RouterName: routerName,
		},
	}
}

// RouterRouteRegisteredEvent is emitted when a route is added to a router.
// The Route field is typed as the event.Route interface (Method + Path)
// so that any route implementation satisfies it regardless of handler type.
type RouterRouteRegisteredEvent struct {
	RouterEvent
	Route Route
}

// NewRouterRouteRegisteredEvent constructs a route registration event.
func NewRouterRouteRegisteredEvent(ctx context.Context, routerName string, rt Route) RouterRouteRegisteredEvent {
	return RouterRouteRegisteredEvent{
		RouterEvent: RouterEvent{
			BaseEvent: BaseEvent{
				eventType: EventTypeRouterRouteRegistered,
				ctx:       ctx,
			},
			RouterName: routerName,
		},
		Route: rt,
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
		RouterEvent: RouterEvent{
			BaseEvent: BaseEvent{
				eventType: EventTypeRouterRequestIncoming,
				ctx:       ctx,
				source:    routerName,
			},
			RouterName: routerName,
		},
		Request:        req,
		ResponseWriter: rw,
	}
}

// RouterRequestHandledEvent records metadata about a completed request.
type RouterRequestHandledEvent struct {
	RouterEvent
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Duration       time.Duration
}

// NewRouterRequestHandledEvent creates an event for a handled request.
func NewRouterRequestHandledEvent(ctx context.Context, routerName string, req *http.Request, rw http.ResponseWriter, dur time.Duration) RouterRequestHandledEvent {
	return RouterRequestHandledEvent{
		RouterEvent: RouterEvent{
			BaseEvent: BaseEvent{
				eventType: EventTypeRouterRequestHandled,
				ctx:       ctx,
				source:    routerName,
			},
			RouterName: routerName,
		},
		Request:        req,
		ResponseWriter: rw,
		Duration:       dur,
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
		RouterEvent: RouterEvent{
			BaseEvent: BaseEvent{
				eventType: EventTypeRouterUnresolvedRequest,
				ctx:       ctx,
				source:    routerName,
			},
			RouterName: routerName,
		},
		Method: method,
	}
}
