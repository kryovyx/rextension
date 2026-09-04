// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the dependency-injection contract from rextension/di,
// so extension code can write rextension.Resolver rather than importing the
// subpackage. See that package for why the declarations live there.
package rextension

import "github.com/kryovyx/rextension/di"

// Resolver resolves dependencies into a provided target.
// Aliased from rextension/di.
type Resolver = di.Resolver

// Scope is a bounded lifetime for scoped dependencies.
// Aliased from rextension/di.
type Scope = di.Scope

// Container manages dependency registration and resolution.
// Aliased from rextension/di.
type Container = di.Container
