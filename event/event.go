// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package event defines the event interfaces, constants, and concrete event types
// used by the Rex framework. Extensions should import this package directly rather
// than depending on the full rex implementation module.
package event

import "context"

// Event represents a generic event in the Rex framework.
type Event interface {
	// Type returns the string identifier for this event type.
	Type() string
	// Context returns the context associated with this event.
	Context() context.Context
}

// EventHandler is the callback signature for event subscriptions.
type EventHandler func(Event)

// EventBus is the event bus interface exposed to extensions and used internally.
type EventBus interface {
	// Subscribe registers a handler for the given event type.
	Subscribe(eventType string, handler EventHandler)
	// Emit publishes an event to all subscribed handlers.
	Emit(event Event)
	// SetLogger configures the bus logger.
	SetLogger(logger BusLogger)
	// Close shuts down the bus.
	Close()
}

// BusLogger is the minimal logger interface required by EventBus.SetLogger.
// Any concrete logger that satisfies rextension.Logger also satisfies this interface,
// keeping the event subpackage free of circular imports.
type BusLogger interface {
	Error(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// Router event type constants using the dot-delimited naming convention.
// These are the canonical event type strings shared across all Rex modules.
const (
	// RouterEventType identifies generic router events.
	RouterEventType = "router_event"
	// EventTypeRouterInitialized is emitted when a router trie is built.
	EventTypeRouterInitialized = "router.initialized"
	// EventTypeRouterRouteRegistered is emitted when a route is added to a router.
	EventTypeRouterRouteRegistered = "router.route.registered"
	// EventTypeRouterRequestIncoming is emitted when an HTTP request arrives.
	EventTypeRouterRequestIncoming = "router.request.incoming"
	// EventTypeRouterRequestHandled is emitted when a request finishes handling.
	EventTypeRouterRequestHandled = "router.request.handled"
	// EventTypeRouterUnresolvedRequest is emitted when no route matches a request.
	EventTypeRouterUnresolvedRequest = "router.request.unresolved"
)

// As attempts to cast the event to the specified type T.
func As[T any](e Event) (T, bool) {
	v, ok := e.(T)
	return v, ok
}
