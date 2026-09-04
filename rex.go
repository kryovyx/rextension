// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the Rex interface, Option type, RouterConfig struct, and related
// constants that extensions need to interact with the framework.
package rextension

import (
	"crypto/tls"
	"reflect"
	"time"

	"github.com/kryovyx/rextension/event"
)

// DefaultRouterName is the name of the default router.
const DefaultRouterName = "default"

// RouterConfig holds configuration for an individual router/listener.
//
// The fields here carried `default:"..."` struct tags, which nothing read
// (D22). They were documentation pretending to be behaviour: the actual
// defaults live in the NewDefaultConfig functions and in the resolve* helpers
// in the router, and the tags had already drifted out of step with them —
// SSLVerify was tagged `default:"true"` while being entirely without effect,
// and ListenSSL was tagged `default:"true"` while defaulting to false.
//
// Defaults are now stated in each field's own comment, next to the field, so
// there is one place to read and one place to change.
//
// # SSLVerify is gone (D26)
//
// The field is removed, not deprecated. It configured nothing: the value was
// stored on the router and never read, and there was no code path that could
// have used it.
//
// It is worth being explicit about what it was mistaken for, because the name
// invited the mistake more than once. `InsecureSkipVerify` is read by
// tls.Client, never tls.Server — so a *listener* has no such setting, and a
// field named SSLVerify on a listener config cannot mean what it appears to.
//
// Client certificate verification, which is what people reached for it
// expecting, is configured through TLSConfig:
//
//	TLSConfig: &tls.Config{
//	    ClientAuth: tls.RequireAndVerifyClientCert,
//	    ClientCAs:  pool,
//	}
//
// That already worked and is already documented below. SSLVerify was a second
// door onto the same room, and it did not open.
type RouterConfig struct {
	// Addr is the address to listen on (e.g., ":8080").
	Addr string
	// BaseURL is the base path prefix for all routes (e.g., "/").
	BaseURL string
	// ListenSSL toggles TLS mode for the listener when cert files are provided.
	ListenSSL bool
	// CertFile is the path to the TLS certificate file (nil disables TLS).
	CertFile *string
	// KeyFile is the path to the TLS key file (nil disables TLS).
	KeyFile *string
	// TLSConfig, when non-nil, is used verbatim for the listener. It takes precedence
	// over CertFile/KeyFile and enables per-handshake certificate selection.
	//
	// Set GetCertificate to swap certificates without a restart: the listener then
	// calls it once per handshake instead of reading a file once at start.
	//
	// This is also where client certificate verification belongs
	// (ClientAuth: tls.RequireAndVerifyClientCert plus ClientCAs) — not SSLVerify.
	TLSConfig *tls.Config

	// ---------------------------------------------------------------------
	// Listener limits
	//
	// net/http applies no timeouts and no body limit of its own. Left unset,
	// a single client holding a connection open without completing its
	// request headers occupies a goroutine indefinitely — Slowloris — and a
	// request body is read until the client stops sending. These fields exist
	// so the defaults are safe rather than merely convenient; every one of
	// them can be raised or disabled per router.
	//
	// Zero means "use the framework default" for the durations and for
	// MaxHeaderBytes. To disable a timeout deliberately, set it negative.
	// ---------------------------------------------------------------------

	// ReadHeaderTimeout bounds the time allowed to read request headers.
	//
	// This is the field that closes Slowloris: the attack works by trickling
	// headers, so the request never completes and the connection is never
	// released. Default 10s.
	ReadHeaderTimeout time.Duration

	// ReadTimeout bounds the time allowed to read the entire request,
	// headers and body. Default 30s.
	//
	// Raise it for routers that accept large uploads over slow links; the
	// per-route body limit is a better tool for size, this one is about time.
	ReadTimeout time.Duration

	// WriteTimeout bounds the time allowed to write the response.
	//
	// Deliberately defaults to 0 — unset — and no framework default fills it
	// in. WriteTimeout is an absolute deadline on the whole response, not an
	// idle timeout, so any non-zero value truncates responses that are
	// legitimately long-lived: server-sent events, long polling, large file
	// downloads, and any slow client on a fast endpoint. Slowloris is a read
	// attack and is already closed by ReadHeaderTimeout, so setting this
	// buys no protection that is not already in place.
	//
	// Set it only on a router you know serves nothing streaming.
	WriteTimeout time.Duration

	// IdleTimeout bounds how long a keep-alive connection may sit unused
	// before it is closed. Default 120s.
	IdleTimeout time.Duration

	// MaxHeaderBytes caps the total size of request headers. Default 1 MiB.
	MaxHeaderBytes int

	// MaxHeaderValueCount caps the *number* of header values in a request.
	// Zero takes net/http's own default of 500; negative also takes it.
	//
	// The count-based sibling of MaxHeaderBytes, and it catches what a byte
	// cap does not: a few thousand tiny headers stay well inside 1 MiB while
	// still forcing the server to allocate and hash every one of them.
	//
	// net/http applies its 500 default whether or not this is set, so leaving
	// it zero is already safe. The field exists so the limit is visible and
	// adjustable alongside the other five rather than being the one bound
	// nothing in the configuration mentions.
	MaxHeaderValueCount int

	// MaxBodyBytes caps the size of a request body, in bytes. Default 4 MiB.
	//
	// Exceeding it fails the read rather than buffering without limit. A
	// route may override this — including to remove the cap — by implementing
	// BodyLimitedRoute.
	//
	// Negative disables the cap for the whole router.
	MaxBodyBytes int64
}

// Listener limit defaults, applied when a RouterConfig leaves the field zero.
//
// WriteTimeout is absent on purpose: see RouterConfig.WriteTimeout.
const (
	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultReadTimeout       = 30 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultMaxHeaderBytes    = 1 << 20 // 1 MiB
	DefaultMaxBodyBytes      = 4 << 20 // 4 MiB

	// DefaultMaxHeaderValueCount mirrors net/http.DefaultMaxHeaderValueCount.
	//
	// Declared here so the framework's own documentation states the effective
	// limit rather than pointing at a constant in another package — but it is
	// not applied by the framework: leaving RouterConfig.MaxHeaderValueCount
	// zero lets net/http apply the same number itself.
	DefaultMaxHeaderValueCount = 500
)

// BodyLimitedRoute is implemented by routes that need a body limit other than
// their router's.
//
// An upload endpoint raises it; a streaming ingest endpoint disables it. The
// value is read once, when the route table is built, not per request.
type BodyLimitedRoute interface {
	// MaxBodyBytes returns the cap in bytes.
	//
	//	 0 — use the router's limit
	//	-1 — no limit
	MaxBodyBytes() int64
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
	//
	// Typed as rextension.Container rather than dix.Container, so an
	// extension needs only this module — and so the container itself is
	// swappable (D23).
	Container() Container
	// EventBus returns the global event bus.
	EventBus() event.EventBus
	// Use registers a standard HTTP middleware on every router, current and
	// future, at PriorityDefault.
	Use(mw Middleware)
	// UseOnRouter registers a middleware on one named router only.
	//
	// The router need not exist yet: the name is resolved when the route table
	// is built, and an unknown name is reported from Run rather than being
	// silently ignored.
	UseOnRouter(routerName string, mw Middleware, priority int)
	// UsePerRoute registers a factory consulted once per route, when the route
	// table is built.
	//
	// This is how an extension attaches middleware to only the routes it
	// applies to. See PerRouteMiddleware.
	UsePerRoute(f PerRouteMiddleware, priority int)
	// UsePerRouter registers a factory consulted once per router, when the
	// route tables are built.
	//
	// For middleware whose configuration depends on which router it runs on.
	// See PerRouterMiddleware.
	UsePerRouter(f PerRouterMiddleware, priority int)
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
