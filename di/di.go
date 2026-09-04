// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package di re-exports the dependency-injection contract from corex/di.
//
// The declarations moved to corex because WSX needs the same three interfaces
// for a per-*connection* scope, which is REX's per-request scope with a longer
// lifetime (W1/W22). `dix` satisfies them unchanged and is not forked.
//
// This package remains as an alias for two reasons, and the first is
// load-bearing:
//
//   - `rex` imports it directly, in route/context_default.go. Removing the
//     package would break a published module at its import line rather than
//     at a symbol, which is the one kind of breakage an alias cannot soften.
//   - rextension/route needs these types, and rextension imports
//     rextension/route, so a declaration in either would be an import cycle.
//     A leaf package is what resolves that, here as in corex.
//
// The aliases are transitive, so rextension.Resolver, rextension/di.Resolver
// and corex/di.Resolver are all one type.
package di

import "github.com/kryovyx/corex/di"

// Resolver resolves dependencies into a provided target.
//
// This is the interface most extension code wants: a health check, a route
// handler and a factory all need to *read* dependencies, not register them.
type Resolver = di.Resolver

// Scope is a bounded lifetime for scoped dependencies.
//
// A scope must be closed to release what it built. The framework creates one
// per request — or, in WSX, one per connection — and closes it when that ends.
type Scope = di.Scope

// Container manages dependency registration and resolution.
//
// Extensions normally want Resolver instead. Container is for the narrow case
// of an extension that publishes something for others to resolve.
type Container = di.Container
