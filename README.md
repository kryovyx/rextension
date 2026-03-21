# Rex Extension Interfaces (rextension)

Minimal interface contract for Rex framework extensions — depend on this instead of the full `rex` module.

[![Go Version](https://img.shields.io/badge/go-1.26+-blue.svg)](https://golang.org/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen.svg)](#)
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
    Container() dix.Container
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
    Addr      string  // Listen address (e.g., ":8080")
    BaseURL   string  // Base path prefix (e.g., "/")
    SSLVerify bool    // Enable SSL certificate verification
    ListenSSL bool    // Toggle TLS mode
    CertFile  *string // Path to TLS certificate file
    KeyFile   *string // Path to TLS key file
}
```

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
| `rextension` | Interface contracts for extensions | `dix` only |
| `rex` | Full framework implementation | `rextension`, `dix` |
| `rextension-*` | Extension implementations | `rextension`, `dix` (not `rex`) |

The `rex` module re-exports all `rextension` types as aliases (e.g., `rex.Extension = rextension.Extension`), so application code that already imports `rex` continues to work unchanged.

## Best Practices

1. **Depend on `rextension`, not `rex`**: Extension modules should only import `rextension` and `dix` to keep the dependency graph minimal
2. **Implement all five hooks**: Even if a hook is a no-op, provide an implementation that returns `nil`
3. **Use `WithExtension` helpers**: Expose a `WithMyExtension()` function returning `rextension.Option` for ergonomic registration
4. **Use the security interfaces**: If your extension deals with auth, implement `SecuritySchemeAccessor` / `SecuredRouteAccessor` to enable cross-extension OpenAPI documentation
5. **Prefer `CreateRouter`**: Use dedicated routers for operational endpoints (health, metrics) to keep them separate from application traffic

## Contributing

**At this time, this project is in active development and is not open for external contributions.** The framework is still being refined and major interfaces may change.

Once the framework reaches a stable architecture and API, contributions from the community will be welcome. Please check back later or open an issue if you have feature requests or feedback.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Copyright

© 2026 Kryovyx
