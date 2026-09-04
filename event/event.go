// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package event defines the event interfaces, constants, and concrete event types
// used by the Rex framework. Extensions should import this package directly rather
// than depending on the full rex implementation module.
//
// # What moved and what stayed (W22)
//
// The transport-neutral half — what an event is, how a handler is called, what
// a bus must offer — is in corex/event, and re-exported here so an extension's
// import line does not change. The *router* events stay: an event type is a
// statement about one framework's lifecycle, and WSX's gateway.* and conn.*
// events belong to WSX for the same reason.
//
// The first design had each framework declare its own structurally identical
// EventBus and rely on Go's structural typing to keep them interchangeable.
// That holds right up until one side gains a method, at which point the module
// that did not change still compiles. One declaration removes the question.
package event

import (
	"context"

	"github.com/kryovyx/corex/event"
)

// Event represents a generic event in the Rex framework. Aliased from
// corex/event.
type Event = event.Event

// EventHandler is the callback signature for event subscriptions. Aliased from
// corex/event.
type EventHandler = event.EventHandler

// EventBus is the event bus interface exposed to extensions and used
// internally. Aliased from corex/event.
type EventBus = event.EventBus

// DropCounter is implemented by an EventBus that discards events under
// saturation and counts how many. Aliased from corex/event.
//
// Consumers should type-assert rather than require it:
//
//	if dc, ok := bus.(event.DropCounter); ok {
//	    gauge.Set(float64(dc.Dropped()))
//	}
//
// Corollary: anything that must be exact cannot be derived from bus events. An
// in-flight request gauge can lose an increment and its matching decrement
// independently, so it belongs in middleware, not in an event subscriber.
type DropCounter = event.DropCounter

// BusLogger is the minimal logger interface required by EventBus.SetLogger.
// Aliased from corex/event. Any concrete logger that satisfies
// rextension.Logger also satisfies it.
type BusLogger = event.BusLogger

// BaseEvent carries the common fields shared by all events. Aliased from
// corex/event.
//
// Its fields are unexported and now live in another module, so a concrete
// event composes it through NewBaseEvent rather than by naming them in a
// composite literal. Every constructor in router_events.go does exactly that;
// see the note there.
type BaseEvent = event.BaseEvent

// NewBaseEvent creates a BaseEvent with the given type, source, and context.
func NewBaseEvent(ctx context.Context, eventType, source string) BaseEvent {
	return event.NewBaseEvent(ctx, eventType, source)
}

// As attempts to cast the event to the specified type T.
//
// A wrapper rather than an alias: a generic function cannot be assigned to a
// variable and has no alias form, so the one line is the only way to
// re-export it.
func As[T any](e Event) (T, bool) {
	return event.As[T](e)
}

// Router event type constants using the dot-delimited naming convention.
// These are the canonical event type strings shared across all Rex modules.
//
// They stay in this module: they name router lifecycle moments, and WSX's
// gateway.* and conn.* constants name its own. A shared constant block would
// mean one framework's vocabulary listing the other's events.
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
