// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports the body-schema contract from corex.
//
// It moved because the reflection schema generator moved with it (W23):
// OpenAPI 3.1 and AsyncAPI 3.0 share JSON Schema 2020-12, so one generator
// serves both documents, and it needs the composition strategies declared
// somewhere both can see. WSX's MessageSchema is the same four kinds over a
// message payload instead of a request body.
//
// Its original reason for existing is unchanged: the OpenAPI generator reads a
// route's schemas through a named interface, rather than with
// reflect.MethodByName("RequestBody") and an unchecked type assertion to
// []interface{} that panicked inside document generation (D20).
package rextension

import "github.com/kryovyx/corex"

// SchemaKind describes the composition strategy for a body schema.
// Aliased from corex.
type SchemaKind = corex.SchemaKind

// Composition strategies, re-declared by value.
const (
	// SchemaScalar indicates a single concrete type.
	SchemaScalar = corex.SchemaScalar
	// SchemaOneOf indicates exactly one of the listed types must match.
	SchemaOneOf = corex.SchemaOneOf
	// SchemaAnyOf indicates one or more of the listed types may match.
	SchemaAnyOf = corex.SchemaAnyOf
	// SchemaAllOf indicates all of the listed types must match (merged).
	SchemaAllOf = corex.SchemaAllOf
)

// BodySchema describes the shape of a request or response body.
// Aliased from corex.
type BodySchema = corex.BodySchema

// BodySchemaProvider is implemented by a route that declares its request and
// response body schemas. Aliased from corex.
type BodySchemaProvider = corex.BodySchemaProvider

// Scalar creates a BodySchema for a single concrete type.
func Scalar(v interface{}) BodySchema { return corex.Scalar(v) }

// OneOf creates a BodySchema where exactly one of the given types must match.
func OneOf(vs ...interface{}) BodySchema { return corex.OneOf(vs...) }

// AnyOf creates a BodySchema where one or more of the given types may match.
func AnyOf(vs ...interface{}) BodySchema { return corex.AnyOf(vs...) }

// AllOf creates a BodySchema where all of the given types must match (merged).
func AllOf(vs ...interface{}) BodySchema { return corex.AllOf(vs...) }
