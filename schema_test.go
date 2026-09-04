// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package rextension_test

import (
	"testing"

	rx "github.com/kryovyx/rextension"
)

type reqA struct{ A string }
type reqB struct{ B int }

func TestSchemaKind_String(t *testing.T) {
	cases := []struct {
		kind rx.SchemaKind
		want string
	}{
		{rx.SchemaScalar, "scalar"},
		{rx.SchemaOneOf, "oneOf"},
		{rx.SchemaAnyOf, "anyOf"},
		{rx.SchemaAllOf, "allOf"},
		{rx.SchemaKind(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("SchemaKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

// The constructors are the entire public surface for building a schema, and
// both the validation middleware and the OpenAPI generator branch on Kind().
func TestBodySchema_constructors(t *testing.T) {
	cases := []struct {
		name      string
		schema    rx.BodySchema
		wantKind  rx.SchemaKind
		wantTypes int
	}{
		{"Scalar", rx.Scalar(reqA{}), rx.SchemaScalar, 1},
		{"OneOf", rx.OneOf(reqA{}, reqB{}), rx.SchemaOneOf, 2},
		{"AnyOf", rx.AnyOf(reqA{}, reqB{}), rx.SchemaAnyOf, 2},
		{"AllOf", rx.AllOf(reqA{}, reqB{}), rx.SchemaAllOf, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.schema.Kind(); got != tc.wantKind {
				t.Fatalf("Kind() = %v, want %v", got, tc.wantKind)
			}
			if got := len(tc.schema.Types()); got != tc.wantTypes {
				t.Fatalf("len(Types()) = %d, want %d", got, tc.wantTypes)
			}
		})
	}
}

// Scalar's documented invariant: exactly one type, and it is the one given.
func TestScalar_carries_the_given_type(t *testing.T) {
	s := rx.Scalar(reqA{A: "x"})
	types := s.Types()
	if len(types) != 1 {
		t.Fatalf("len(Types()) = %d, want 1", len(types))
	}
	v, ok := types[0].(reqA)
	if !ok {
		t.Fatalf("Types()[0] is %T, want reqA", types[0])
	}
	if v.A != "x" {
		t.Fatalf("value not preserved: %+v", v)
	}
}

// A union constructor called with no arguments produces an empty type list
// rather than panicking. Consumers branch on Kind() and would otherwise index
// into nothing, so pinning the behaviour matters more than the value of it.
func TestUnion_constructors_with_no_types(t *testing.T) {
	for _, s := range []rx.BodySchema{rx.OneOf(), rx.AnyOf(), rx.AllOf()} {
		if len(s.Types()) != 0 {
			t.Fatalf("expected no types, got %d", len(s.Types()))
		}
	}
}
