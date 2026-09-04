# Rex Extension Interfaces (rextension)

Minimal interface contract for Rex framework extensions — depend on this instead of the full `rex` module.

[![Go Version](https://img.shields.io/badge/go-1.27+-blue.svg)](https://golang.org/dl/)
[![Coverage](https://img.shields.io/badge/coverage-94.9%25-brightgreen.svg)](#)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Overview

`rextension` defines the canonical interfaces that Rex extensions implement and interact with. By depending on this lightweight module instead of the full `rex` implementation, extension authors avoid pulling in the entire framework as a dependency.

This module provides:

- **Rex interface**: The subset of the framework API available to extensions
- **Extension interface**: Five lifecycle hooks for application customization
- **Logger interface**: Logging abstraction with configurable log levels
- **EventBus interface**: Event subscription and emission
- **Route interface**: Minimal route definition (method + path)
- **Middleware type**: Standard Go `func(http.Handler) http.Handler` middleware
- **RouterConfig struct**: Configuration for routers (address, TLS, base URL)
- **Option type**: Functional option for configuring Rex instances
- **Security interfaces**: `SecuritySchemeAccessor` and `SecuredRouteAccessor` for cross-extension security contracts
- **Helper functions**: `WithExtension` / `WithExtensions` for registering extensions
- **Global security registry**: `RegisterSecuritySchemes` / `GetSecuritySchemes` (concurrent-safe)

## Installation

```bash
go get github.com/kryovyx/rextension
```

## Interfaces

### Rex

The core interface extensions receive in their lifecycle callbacks:

```go
type Rex interface {
    Logger() Logger
    Container() Container   // the DI contract declared in this module
    EventBus() EventBus
    Use(mw Middleware)
    RegisterRoute(rt Route) error
    RegisterRouteToRouter(rt Route, routerName string) error
    CreateRouter(name string, cfg RouterConfig) error
}
```

### Extension

Implement this interface to create a Rex extension. Five lifecycle hooks are called in order:

```go
type Extension interface {
    OnInitialize(ctx context.Context, r Rex) error  // Register routes, subscribe to events
    OnStart(ctx context.Context, r Rex) error       // Start background work
    OnReady(ctx context.Context, r Rex) error       // All listeners are up
    OnStop(ctx context.Context, r Rex) error        // Application is stopping
    OnShutdown(ctx context.Context, r Rex) error    // All resources released
}
```

### Logger

```go
type LogLevel int

const (
    LogLevelTrace LogLevel = iota
    LogLevelDebug
    LogLevelInfo
    LogLevelWarn
    LogLevelError
    LogLevelOff
)

type Logger interface {
    Info(format string, args ...interface{})
    Warn(format string, args ...interface{})
    Error(format string, args ...interface{})
    Debug(format string, args ...interface{})
    Trace(format string, args ...interface{})
    SetLogLevel(level LogLevel)
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithError(err error) Logger
}
```

### EventBus

```go
type Event interface {
    Type() string
    Context() context.Context
}

type EventHandler func(Event)

type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Emit(event Event)
    SetLogger(logger Logger)
    Close()
}
```

### Route

```go
type Route interface {
    Method() string
    Path() string
}
```

### Middleware

The standard Go HTTP middleware type:

```go
type Middleware func(http.Handler) http.Handler
```

### RouterConfig

```go
type RouterConfig struct {
    Addr      string      // Listen address (e.g., ":8080")
    BaseURL   string      // Base path prefix (e.g., "/")
    ListenSSL bool        // Toggle TLS mode
    CertFile  *string     // Path to TLS certificate file
    KeyFile   *string     // Path to TLS key file
    TLSConfig *tls.Config // Takes precedence over CertFile/KeyFile

    // Listener limits. Zero takes the default; negative disables.
    ReadHeaderTimeout time.Duration // default 10s — the Slowloris bound
    ReadTimeout       time.Duration // default 30s
    WriteTimeout      time.Duration // default 0 — unset, deliberately
    IdleTimeout       time.Duration // default 120s
    MaxHeaderBytes    int           // default 1 MiB
    MaxBodyBytes      int64         // default 4 MiB
}
```

`TLSConfig` is the injection point for a caller-supplied `*tls.Config`. When it is set,
the listener uses it verbatim and ignores `CertFile`/`KeyFile`. Setting `GetCertificate`
makes the certificate a per-handshake decision, which is what allows a certificate to be
replaced without restarting the process.

**Client certificate verification** is configured through `TLSConfig`:

```go
TLSConfig: &tls.Config{
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs:  pool,
}
```

### Listener limits

`net/http` applies no timeouts and no body limit of its own, so every one of these
defaults to "unlimited" unless something sets it. Left unset, a single client holding a
connection open without completing its request headers occupies a goroutine for as long
as it cares to — Slowloris — and a request body is read until the client stops sending.

`ReadHeaderTimeout` is the field that closes Slowloris, because the attack works by
trickling headers so the request never completes.

**`WriteTimeout` defaults to 0 and is never filled in.** It is an absolute deadline on
the whole response, not an idle timeout, so any non-zero value truncates responses that
are legitimately long-lived: server-sent events, long polling, large downloads, and any
slow client on a fast endpoint. Slowloris is a *read* attack and is already closed by
`ReadHeaderTimeout`, so setting this buys no protection that is not already in place.
Set it only on a router you know serves nothing streaming.

A route may override the body limit by implementing `BodyLimitedRoute`:

```go
func (r *UploadRoute) MaxBodyBytes() int64 { return 64 << 20 } // 64 MiB
func (r *IngestRoute) MaxBodyBytes() int64 { return -1 }       // no limit
```

### `SSLVerify` was removed

The field configured nothing: the value was stored on the router and never read, and
there was no code path that could have used it.

It is worth saying what it was mistaken for, because the name invited the mistake more
than once. `InsecureSkipVerify` is read by `tls.Client`, never `tls.Server` — so a
*listener* has no such setting, and a field named `SSLVerify` on a listener config
cannot mean what it appears to. Client certificate verification, which is what people
reached for it expecting, is `TLSConfig.ClientAuth` above. That always worked; `SSLVerify`
was a second door onto the same room, and it did not open.

### SecuritySchemeAccessor

Allows security extensions to expose scheme metadata for OpenAPI documentation without import coupling:

```go
type SecuritySchemeAccessor interface {
    Name() string        // Unique identifier (e.g., "bearer", "basic")
    Type() string        // OpenAPI type (e.g., "http", "apiKey")
    Description() string // Human-readable description
    Challenge() string   // WWW-Authenticate value (e.g., "Bearer")
}
```

### SecuredRouteAccessor

Allows routes to declare which security schemes they require:

```go
type SecuredRouteAccessor interface {
    RequiredSchemes() []string // Empty/nil means public
}
```

## Constants

```go
const DefaultRouterName = "default"
```

## Helper Functions

Register extensions via functional options without importing the full `rex` module:

```go
// Single extension
opt := rextension.WithExtension(myExtension)

// Multiple extensions
opt := rextension.WithExtensions(ext1, ext2, ext3)
```

## Global Security Registry

A concurrent-safe, package-level registry for sharing security schemes between extensions (e.g., the security extension publishes schemes, the OpenAPI extension reads them):

```go
// Register schemes (typically called in OnInitialize or OnStart)
rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{
    myBearerScheme,
    myAPIKeyScheme,
})

// Retrieve a snapshot of registered schemes
schemes := rextension.GetSecuritySchemes() // returns nil if none registered
```

## Writing an Extension

A minimal extension using only the `rextension` module:

```go
package myext

import (
    "context"
    rx "github.com/kryovyx/rextension"
)

type MyExtension struct{}

func NewMyExtension() rx.Extension {
    return &MyExtension{}
}

func WithMyExtension() rx.Option {
    return rx.WithExtension(NewMyExtension())
}

func (e *MyExtension) OnInitialize(ctx context.Context, r rx.Rex) error {
    r.Logger().Info("MyExtension initializing")
    return nil
}

func (e *MyExtension) OnStart(ctx context.Context, r rx.Rex) error  { return nil }
func (e *MyExtension) OnReady(ctx context.Context, r rx.Rex) error  { return nil }
func (e *MyExtension) OnStop(ctx context.Context, r rx.Rex) error   { return nil }
func (e *MyExtension) OnShutdown(ctx context.Context, r rx.Rex) error { return nil }
```

## Relationship to rex

| Module | Purpose | Depends on |
|--------|---------|------------|
| `rextension` | Interface contracts for extensions | **nothing** |
| `rex` | Full framework implementation | `rextension`, `dix` |
| `rextension-*` | Extension implementations | `rextension` only (not `rex`, not `dix`) |

The `rex` module re-exports all `rextension` types as aliases (e.g., `rex.Extension = rextension.Extension`), so application code that already imports `rex` continues to work unchanged.

## Best Practices

1. **Depend on `rextension`, not `rex`**: an extension module should import
   `rextension` and nothing else from this ecosystem. It does not need `dix`
   either — the dependency-injection contract (`Container`, `Resolver`,
   `Scope`) is declared here and satisfied by `dix`, so an extension asks for
   `rextension.Resolver` and never names the implementation
2. **Implement all five hooks**: Even if a hook is a no-op, provide an implementation that returns `nil`
3. **Use `WithExtension` helpers**: Expose a `WithMyExtension()` function returning `rextension.Option` for ergonomic registration
4. **Use the security interfaces**: If your extension deals with auth, implement `SecuritySchemeAccessor` / `SecuredRouteAccessor` to enable cross-extension OpenAPI documentation
5. **Prefer `CreateRouter`**: Use dedicated routers for operational endpoints (health, metrics) to keep them separate from application traffic

## Contributing

**The framework is in alpha, and external contributions open at `v1.0.0`.**
Until then pull requests will be closed unmerged — but issues are very welcome.
Bug reports, questions and feature requests all feed into what `v1.0.0` looks
like.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the rules that will apply, and
[COMMIT-CONVENTIONS.md](COMMIT-CONVENTIONS.md) for the commit format.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Copyright

© 2026 Kryovyx
