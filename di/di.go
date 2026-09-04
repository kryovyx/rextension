// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package di declares the dependency-injection contract that Rex extensions
// depend on.
//
// These interfaces mirror the ones in github.com/kryovyx/dix, which implements
// them. They exist here so that an extension needs only the rextension module:
// `dix` currently appears in the go.mod of six extensions that import it in
// **zero** Go files, purely because the Rex interface hands back a
// dix.Container (D23).
//
// The practical consequence, beyond a tidier module graph, is that a change to
// dix's Container interface stops reaching extension code — including their
// test mocks, which is exactly the churn adding dix.Unbind caused.
//
// # Why its own package
//
// It is a leaf: it imports nothing. Both rextension and rextension/route need
// these types, and rextension already imports rextension/route (for
// PerRouteMiddleware), so declaring them in either one would be an import
// cycle. rextension re-exports all three as aliases, so extension code can
// keep writing rextension.Resolver.
package di

// Resolver resolves dependencies into a provided target.
//
// This is the interface most extension code wants: a health check, a route
// handler and a factory all need to *read* dependencies, not register them.
type Resolver interface {
	// Resolve sets *target to the resolved dependency. target must be a
	// pointer.
	Resolve(target any) error

	// ResolveAll appends every resolvable dependency to *target, which must be
	// a pointer to a slice of interfaces.
	ResolveAll(target any) error
}

// Scope is a bounded lifetime for scoped dependencies.
//
// A scope must be closed to release what it built. The framework creates one
// per request and closes it when the handler returns.
type Scope interface {
	Resolver

	// Close releases every scoped instance that has a Close method.
	Close() error
}

// Container manages dependency registration and resolution.
//
// Extensions normally want Resolver instead. Container is for the narrow case
// of an extension that publishes something for others to resolve.
type Container interface {
	Resolver

	// Singleton registers a factory instantiated once, on first resolve, and
	// shared. Accepts func() T or func(Resolver) T.
	Singleton(factory any) error

	// Scoped registers a factory instantiated once per scope.
	Scoped(factory any) error

	// Transient registers a factory instantiated on every resolve.
	Transient(factory any) error

	// Instance registers a pre-constructed value.
	Instance(v any) error

	// Unbind removes the registration for the exact type of v, reporting
	// whether anything was removed. It is how a registration is corrected: a
	// type may hold only one, and the replacement is often a different
	// concrete type than the original.
	Unbind(v any) (bool, error)
}

// Scope creation is deliberately absent from Container.
//
// Creating and closing a scope is the framework's job: the router opens one per
// request and closes it when the handler returns. No extension in the ecosystem
// calls it — the only implementations of a NewScope method were the test mocks
// each extension had to write in order to satisfy an interface it never used.
//
// Leaving it out also means dix.Container satisfies this interface directly.
// Including it would not: dix.NewScope returns dix.Scope, Go interfaces are not
// covariant, and dix cannot import this package without a module cycle. The
// alternative was an adapter in the framework whose only purpose was to
// re-type a method nobody called.
