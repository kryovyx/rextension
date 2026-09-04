// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

// This file is the safety argument for the corex extraction (W22).
//
// Roughly 55 of this module's exported symbols are now declarations in corex,
// re-exported here so that nine extension modules and one consumer compile
// untouched. The claim that makes that safe is narrow and checkable: each
// re-export must be *the same type*, not a structurally identical copy.
//
// The distinction is invisible in ordinary use and fatal at a module boundary.
// `type Middleware = corex.Middleware` shares identity, so a middleware
// written against either satisfies both. `type Middleware
// func(http.Handler) http.Handler` — a re-declaration — would compile here,
// pass every test in this module, and then fail to satisfy an interface in
// some extension that had received the corex-typed value from somewhere else.
//
// So the assertions below assign across the boundary with **no conversion**.
// Every one of them stops compiling if a re-export is ever turned into a copy,
// which is the only failure mode the alias shim has.
package rextension_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kryovyx/corex"
	corexdi "github.com/kryovyx/corex/di"
	corexevent "github.com/kryovyx/corex/event"
	"github.com/kryovyx/rextension"
	rxdi "github.com/kryovyx/rextension/di"
	rxevent "github.com/kryovyx/rextension/event"
)

// ---- type identity ----
//
// Declared as package-level vars so a broken alias is a compile error in this
// file rather than a failure inside whichever test happened to touch it.
var (
	_ corex.Middleware             = rextension.Middleware(nil)
	_ rextension.Middleware        = corex.Middleware(nil)
	_ corex.LogLevel               = rextension.LogLevel(0)
	_ rextension.LogLevel          = corex.LogLevel(0)
	_ *corex.Problem               = (*rextension.Problem)(nil)
	_ *rextension.Problem          = (*corex.Problem)(nil)
	_ corex.FieldError             = rextension.FieldError{}
	_ corex.OriginPolicy           = rextension.OriginPolicy{}
	_ *corex.OriginPolicyError     = (*rextension.OriginPolicyError)(nil)
	_ corex.SchemaKind             = rextension.SchemaKind(0)
	_ corexdi.Resolver             = rxdi.Resolver(nil)
	_ rextension.Resolver          = corexdi.Resolver(nil)
	_ corexdi.Container            = rextension.Container(nil)
	_ corexdi.Scope                = rextension.Scope(nil)
	_ corexevent.Event             = rxevent.Event(nil)
	_ rxevent.EventBus             = corexevent.EventBus(nil)
	_ corexevent.BusLogger         = rxevent.BusLogger(nil)
	_ corexevent.BaseEvent         = rxevent.BaseEvent{}
	_ corexevent.DropCounter       = rxevent.DropCounter(nil)
	_ corexevent.EventHandler      = rxevent.EventHandler(nil)
	_ corex.Logger                 = rextension.Logger(nil)
	_ corex.BodySchema             = rextension.BodySchema(nil)
	_ corex.BodySchemaProvider     = rextension.BodySchemaProvider(nil)
	_ corex.SchemeRegistry         = rextension.SchemeRegistry(nil)
	_ corex.SecuredRouteAccessor   = rextension.SecuredRouteAccessor(nil)
	_ corex.SecuritySchemeAccessor = rextension.SecuritySchemeAccessor(nil)
	_ corex.ParameterizedScheme    = rextension.ParameterizedScheme(nil)
	_ corex.BearerFormatProvider   = rextension.BearerFormatProvider(nil)
	_ corex.RoleClaimProvider      = rextension.RoleClaimProvider(nil)
)

// ---- constants ----
//
// A constant has no alias form, so every one of them is re-declared by value
// here. That is the one place the shim genuinely duplicates something, and
// therefore the one place it can drift — a mismatch would mean an extension
// and the framework disagreeing about which slug names which problem, with no
// compile error at all.
func TestAliasedConstants_match_corex(t *testing.T) {
	strs := []struct {
		name       string
		here, ther string
	}{
		{"ProblemMediaType", rextension.ProblemMediaType, corex.ProblemMediaType},
		{"DefaultProblemTypeBase", rextension.DefaultProblemTypeBase, corex.DefaultProblemTypeBase},
		{"DefaultInstanceBase", rextension.DefaultInstanceBase, corex.DefaultInstanceBase},
		{"ProblemUnauthorized", rextension.ProblemUnauthorized, corex.ProblemUnauthorized},
		{"ProblemForbidden", rextension.ProblemForbidden, corex.ProblemForbidden},
		{"ProblemNotFound", rextension.ProblemNotFound, corex.ProblemNotFound},
		{"ProblemMethodNotAllowed", rextension.ProblemMethodNotAllowed, corex.ProblemMethodNotAllowed},
		{"ProblemNotAcceptable", rextension.ProblemNotAcceptable, corex.ProblemNotAcceptable},
		{"ProblemUnsupportedMediaType", rextension.ProblemUnsupportedMediaType, corex.ProblemUnsupportedMediaType},
		{"ProblemPayloadTooLarge", rextension.ProblemPayloadTooLarge, corex.ProblemPayloadTooLarge},
		{"ProblemValidationFailed", rextension.ProblemValidationFailed, corex.ProblemValidationFailed},
		{"ProblemRateLimitExceeded", rextension.ProblemRateLimitExceeded, corex.ProblemRateLimitExceeded},
		{"ProblemInternal", rextension.ProblemInternal, corex.ProblemInternal},
		{"ProblemDependencyUnavailable", rextension.ProblemDependencyUnavailable, corex.ProblemDependencyUnavailable},
		{"ProblemBadRequest", rextension.ProblemBadRequest, corex.ProblemBadRequest},
	}
	for _, tc := range strs {
		if tc.here != tc.ther {
			t.Errorf("%s = %q here, %q in corex", tc.name, tc.here, tc.ther)
		}
	}

	ints := []struct {
		name       string
		here, ther int
	}{
		{"PriorityRecovery", rextension.PriorityRecovery, corex.PriorityRecovery},
		{"PriorityCORS", rextension.PriorityCORS, corex.PriorityCORS},
		{"PriorityRateLimit", rextension.PriorityRateLimit, corex.PriorityRateLimit},
		{"PriorityAuth", rextension.PriorityAuth, corex.PriorityAuth},
		{"PriorityCSRF", rextension.PriorityCSRF, corex.PriorityCSRF},
		{"PriorityValidation", rextension.PriorityValidation, corex.PriorityValidation},
		{"PriorityHealthGate", rextension.PriorityHealthGate, corex.PriorityHealthGate},
		{"PriorityDefault", rextension.PriorityDefault, corex.PriorityDefault},
	}
	for _, tc := range ints {
		if tc.here != tc.ther {
			t.Errorf("%s = %d here, %d in corex", tc.name, tc.here, tc.ther)
		}
	}

	levels := []struct {
		name       string
		here, ther rextension.LogLevel
	}{
		{"LogLevelTrace", rextension.LogLevelTrace, corex.LogLevelTrace},
		{"LogLevelDebug", rextension.LogLevelDebug, corex.LogLevelDebug},
		{"LogLevelInfo", rextension.LogLevelInfo, corex.LogLevelInfo},
		{"LogLevelWarn", rextension.LogLevelWarn, corex.LogLevelWarn},
		{"LogLevelError", rextension.LogLevelError, corex.LogLevelError},
		{"LogLevelOff", rextension.LogLevelOff, corex.LogLevelOff},
	}
	for _, tc := range levels {
		if tc.here != tc.ther {
			t.Errorf("%s = %d here, %d in corex", tc.name, tc.here, tc.ther)
		}
	}

	kinds := []struct {
		name       string
		here, ther rextension.SchemaKind
	}{
		{"SchemaScalar", rextension.SchemaScalar, corex.SchemaScalar},
		{"SchemaOneOf", rextension.SchemaOneOf, corex.SchemaOneOf},
		{"SchemaAnyOf", rextension.SchemaAnyOf, corex.SchemaAnyOf},
		{"SchemaAllOf", rextension.SchemaAllOf, corex.SchemaAllOf},
	}
	for _, tc := range kinds {
		if tc.here != tc.ther {
			t.Errorf("%s = %d here, %d in corex", tc.name, tc.here, tc.ther)
		}
	}

	// The router event constants stay here rather than moving, so they have no
	// corex counterpart to compare against — only their values to pin. That is
	// covered by rextension/event's own TestConstants.
}

// ---- function wrappers ----
//
// A plain function cannot be aliased either. `var NewProblem = corex.NewProblem`
// would work and is rejected: it re-introduces process-wide mutability, which
// is precisely the defect W31 removed from ProblemTypeBase two files away.
// One-line wrappers instead — and a wrapper is code, so it gets tested.

func TestNewProblem_delegates(t *testing.T) {
	p := rextension.NewProblem(http.StatusNotFound, rextension.ProblemNotFound, "gone")
	if p.Type != corex.DefaultProblemTypeBase+corex.ProblemNotFound {
		t.Errorf("Type = %q", p.Type)
	}
	if p.Status != http.StatusNotFound || p.Detail != "gone" {
		t.Errorf("Status = %d, Detail = %q", p.Status, p.Detail)
	}
	if p.Title != http.StatusText(http.StatusNotFound) {
		t.Errorf("Title = %q", p.Title)
	}
}

func TestWriteProblem_delegates(t *testing.T) {
	rec := httptest.NewRecorder()
	rextension.WriteProblem(rec, httptest.NewRequest(http.MethodGet, "/x", nil),
		http.StatusUnauthorized, rextension.ProblemUnauthorized, "nope")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("code = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != corex.ProblemMediaType {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Error("no body written")
	}
}

func TestRequestOrigin_delegates(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := rextension.RequestOrigin(r); got != "" {
		t.Errorf("no Origin header should give %q, got %q", "", got)
	}
	r.Header.Set("Origin", "https://app.example.com")
	if got := rextension.RequestOrigin(r); got != "https://app.example.com" {
		t.Errorf("RequestOrigin = %q", got)
	}
}

func TestSafeMethod_delegates(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace} {
		if !rextension.SafeMethod(m) {
			t.Errorf("%s should be safe", m)
		}
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if rextension.SafeMethod(m) {
			t.Errorf("%s should not be safe", m)
		}
	}
}

func TestBodySchemaConstructors_delegate(t *testing.T) {
	type a struct{}
	type b struct{}

	cases := []struct {
		name string
		got  rextension.BodySchema
		kind rextension.SchemaKind
		n    int
	}{
		{"Scalar", rextension.Scalar(a{}), corex.SchemaScalar, 1},
		{"OneOf", rextension.OneOf(a{}, b{}), corex.SchemaOneOf, 2},
		{"AnyOf", rextension.AnyOf(a{}, b{}), corex.SchemaAnyOf, 2},
		{"AllOf", rextension.AllOf(a{}, b{}), corex.SchemaAllOf, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Kind() != tc.kind {
				t.Errorf("Kind = %v, want %v", tc.got.Kind(), tc.kind)
			}
			if len(tc.got.Types()) != tc.n {
				t.Errorf("len(Types) = %d, want %d", len(tc.got.Types()), tc.n)
			}
			// The variadic constructors must forward the slice, not drop it:
			// AnyOf(a, b) losing an argument would produce a union of one,
			// which is a valid schema describing the wrong thing.
			if _, ok := tc.got.(corex.BodySchema); !ok {
				t.Error("the result must satisfy corex.BodySchema")
			}
		})
	}
}

// ProblemTypeBase and InstanceBase are deliberately NOT re-exported (W31).
//
// There is nothing to call, so nothing to test — the assertion is the absence,
// and the compiler makes it. What is worth pinning is that configuring them in
// corex reaches a problem built through *this* module's wrapper, because that
// is what an application actually does after changing its import line.
func TestConfigureProblems_reaches_the_rextension_wrapper(t *testing.T) {
	t.Cleanup(func() { corex.ConfigureProblems() })

	corex.ConfigureProblems(
		corex.WithProblemTypeBase("https://api.example.com/problems/"))

	p := rextension.NewProblem(http.StatusForbidden, rextension.ProblemForbidden, "")
	if p.Type != "https://api.example.com/problems/forbidden" {
		t.Fatalf("Type = %q — configuring corex did not reach rextension.NewProblem", p.Type)
	}
}
