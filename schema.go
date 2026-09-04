// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file declares the body-schema contract that routes use to describe their
// request and response payloads.
//
// It lives here, rather than in the validation extension, so that the OpenAPI
// generator can read a route's schemas through a named interface. It previously
// reached them with reflect.MethodByName("RequestBody") and
// MethodByName("Responses"), then type-asserted the result to
// []interface{} — unchecked, so an unexpected slice type panicked inside
// document generation. Reflection was the workaround for the two extensions not
// sharing a type; sharing one removes the need (D20).
package rextension

// SchemaKind describes the composition strategy for a body schema.
type SchemaKind int

const (
	// SchemaScalar indicates a single concrete type.
	SchemaScalar SchemaKind = iota
	// SchemaOneOf indicates exactly one of the listed types must match.
	SchemaOneOf
	// SchemaAnyOf indicates one or more of the listed types may match.
	SchemaAnyOf
	// SchemaAllOf indicates all of the listed types must match (merged).
	SchemaAllOf
)

// String returns a human-readable name for the schema kind.
func (k SchemaKind) String() string {
	switch k {
	case SchemaScalar:
		return "scalar"
	case SchemaOneOf:
		return "oneOf"
	case SchemaAnyOf:
		return "anyOf"
	case SchemaAllOf:
		return "allOf"
	default:
		return "unknown"
	}
}

// BodySchema describes the shape of a request or response body.
// It can represent a single type (Scalar) or a union (OneOf/AnyOf/AllOf).
type BodySchema interface {
	// Kind returns the composition strategy.
	Kind() SchemaKind
	// Types returns the zero-value struct(s) that define the schema.
	// For Scalar, len(Types()) == 1. For unions, len >= 2.
	Types() []interface{}
}

// bodySchema is the default implementation of BodySchema.
type bodySchema struct {
	kind  SchemaKind
	types []interface{}
}

func (b *bodySchema) Kind() SchemaKind     { return b.kind }
func (b *bodySchema) Types() []interface{} { return b.types }

var _ BodySchema = (*bodySchema)(nil)

// Scalar creates a BodySchema for a single concrete type.
func Scalar(v interface{}) BodySchema {
	return &bodySchema{kind: SchemaScalar, types: []interface{}{v}}
}

// OneOf creates a BodySchema where exactly one of the given types must match.
func OneOf(vs ...interface{}) BodySchema {
	return &bodySchema{kind: SchemaOneOf, types: vs}
}

// AnyOf creates a BodySchema where one or more of the given types may match.
func AnyOf(vs ...interface{}) BodySchema {
	return &bodySchema{kind: SchemaAnyOf, types: vs}
}

// AllOf creates a BodySchema where all of the given types must match (merged).
func AllOf(vs ...interface{}) BodySchema {
	return &bodySchema{kind: SchemaAllOf, types: vs}
}

// BodySchemaProvider is implemented by a route that declares its request and
// response body schemas.
//
// Both the validation middleware and the OpenAPI generator read routes through
// this interface. A route that does not implement it is passed through without
// validation and documented without schemas.
type BodySchemaProvider interface {
	// RequestBody returns the body schema for the request, or nil if the
	// route does not accept a request body.
	RequestBody() BodySchema

	// Responses returns a map of HTTP status code to body schema.
	// Return nil to skip response validation entirely.
	Responses() map[int]BodySchema
}
