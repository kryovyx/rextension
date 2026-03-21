// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Rex interface, Option type, RouterConfig struct, and related
// constants that extensions need to interact with the framework.
package rextension

import (
	"reflect"

	"github.com/kryovyx/dix"
	"github.com/kryovyx/rextension/event"
)

// DefaultRouterName is the name of the default router.
const DefaultRouterName = "default"

// RouterConfig holds configuration for an individual router/listener.
type RouterConfig struct {
	// Addr is the address to listen on (e.g., ":8080").
	Addr string `default:":8080"`
	// BaseURL is the base path prefix for all routes (e.g., "/").
	BaseURL string `default:"/"`
	// SSLVerify enables SSL certificate verification for outbound connections.
	SSLVerify bool `default:"true"`
	// ListenSSL toggles TLS mode for the listener when cert files are provided.
	ListenSSL bool `default:"true"`
	// CertFile is the path to the TLS certificate file (nil disables TLS).
	CertFile *string `default:"nil"`
	// KeyFile is the path to the TLS key file (nil disables TLS).
	KeyFile *string `default:"nil"`
}

// Option is a functional option for configuring or extending a Rex instance.
// It is the type accepted by rex.New and rex.Rex.WithOptions.
type Option func(r Rex)

// Rex is the interface that extensions interact with during their lifecycle callbacks.
// It exposes the subset of the Rex framework that extensions need.
type Rex interface {
	// Logger returns the global logger.
	Logger() Logger
	// Container returns the root dependency injection container.
	Container() dix.Container
	// EventBus returns the global event bus.
	EventBus() event.EventBus
	// Use registers a standard HTTP middleware on the default router.
	Use(mw Middleware)
	// RegisterRoute registers a route on the default router.
	RegisterRoute(rt Route) error
	// RegisterRouteToRouter registers a route on the named router.
	RegisterRouteToRouter(rt Route, routerName string) error
	// CreateRouter creates a new named router with the given configuration.
	CreateRouter(name string, cfg RouterConfig) error
}

// WithExtension returns an Option that adds the given extension to the Rex instance.
// The Rex value passed to the option must also implement a WithExtensions method;
// the concrete rex.Rex type satisfies this.
func WithExtension(ext Extension) Option {
	return func(r Rex) {
		// Try to call WithExtensions using reflection to handle various return types.
		rv := reflect.ValueOf(r)
		method := rv.MethodByName("WithExtensions")
		if method.IsValid() && method.Type().NumIn() == 1 {
			// Call WithExtensions(ext)
			method.Call([]reflect.Value{reflect.ValueOf(ext)})
		}
	}
}

// WithExtensions returns an Option that adds multiple extensions to the Rex instance.
func WithExtensions(ext ...Extension) Option {
	return func(r Rex) {
		// Try to call WithExtensions using reflection to handle various return types.
		rv := reflect.ValueOf(r)
		method := rv.MethodByName("WithExtensions")
		if method.IsValid() {
			// Convert ext slice to reflect.Value slice for variadic call
			args := make([]reflect.Value, len(ext))
			for i, e := range ext {
				args[i] = reflect.ValueOf(e)
			}
			method.Call(args)
		}
	}
}
