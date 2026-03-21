// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This package provides all the types an extension needs to implement and interact with
// the Rex framework, without depending on the full rex implementation module.
//
// This file declares the Extension interface.
package rextension

import "context"

// Extension represents a Rex framework extension.
// Implement this interface to hook into the Rex application lifecycle.
type Extension interface {
	// OnInitialize is called once when the extension is registered.
	// Use this to set up infrastructure, subscribe to events, and register routes.
	OnInitialize(ctx context.Context, r Rex) error
	// OnStart is called when the Rex application starts.
	OnStart(ctx context.Context, r Rex) error
	// OnReady is called after all listeners have started successfully.
	OnReady(ctx context.Context, r Rex) error
	// OnStop is called when the Rex application is stopping.
	OnStop(ctx context.Context, r Rex) error
	// OnShutdown is called after all resources have been shut down.
	OnShutdown(ctx context.Context, r Rex) error
}
