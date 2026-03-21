// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Event, EventHandler, and EventBus interfaces used by extensions.
package rextension

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
	SetLogger(logger Logger)
	// Close shuts down the bus.
	Close()
}
