// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file re-exports RFC 9457 problem details from corex — the single error
// format for the whole stack (D30/O9), and now for both stacks: a WebSocket
// error payload is an rextension.Problem, which is a concrete argument for the
// bridge module rather than a coincidence (W5).
//
// # One symbol pair is deliberately not re-exported (W31)
//
// ProblemTypeBase and InstanceBase were mutable package-level vars here. They
// are gone, and this file does not replace them, because there is no way to
// replace them that is not worse:
//
//   - `type X = corex.X` does not apply — a variable has no alias form.
//   - `var ProblemTypeBase = corex.ProblemTypeBase` makes a **copy**. An
//     application setting rextension.ProblemTypeBase would then configure
//     nothing at all, silently, with the framework continuing to read its own
//     value. A silent no-op is worse than a compile error, which is the whole
//     argument for breaking here rather than shimming.
//
// They were also process-wide mutable configuration read on every problem
// construction — the shape D21 removed for the security scheme registry — so
// the extraction was a free opportunity to fix an already-diagnosed defect.
//
// An application that set either one calls corex.ConfigureProblems instead:
//
//	corex.ConfigureProblems(
//	    corex.WithProblemTypeBase("https://api.example.com/problems/"))
//
// This is the only breaking symbol in the whole corex extraction.
package rextension

import (
	"net/http"

	"github.com/kryovyx/corex"
)

// ProblemMediaType is the media type RFC 9457 defines.
const ProblemMediaType = corex.ProblemMediaType

// DefaultProblemTypeBase is the prefix for framework problem types.
const DefaultProblemTypeBase = corex.DefaultProblemTypeBase

// DefaultInstanceBase is the prefix for the instance member.
const DefaultInstanceBase = corex.DefaultInstanceBase

// Framework problem type slugs, re-declared by value.
const (
	// ProblemUnauthorized — authentication is required or failed.
	ProblemUnauthorized = corex.ProblemUnauthorized
	// ProblemForbidden — authenticated, but not permitted.
	ProblemForbidden = corex.ProblemForbidden
	// ProblemNotFound — no route matches the request path.
	ProblemNotFound = corex.ProblemNotFound
	// ProblemMethodNotAllowed — the path exists under other methods.
	ProblemMethodNotAllowed = corex.ProblemMethodNotAllowed
	// ProblemNotAcceptable — no representation matches the Accept header.
	ProblemNotAcceptable = corex.ProblemNotAcceptable
	// ProblemUnsupportedMediaType — the request body's Content-Type is not
	// supported.
	ProblemUnsupportedMediaType = corex.ProblemUnsupportedMediaType
	// ProblemPayloadTooLarge — the request body exceeds the configured limit.
	ProblemPayloadTooLarge = corex.ProblemPayloadTooLarge
	// ProblemValidationFailed — the request body failed validation.
	ProblemValidationFailed = corex.ProblemValidationFailed
	// ProblemRateLimitExceeded — the caller has exceeded its rate limit.
	ProblemRateLimitExceeded = corex.ProblemRateLimitExceeded
	// ProblemInternal — an unexpected server-side failure.
	ProblemInternal = corex.ProblemInternal
	// ProblemDependencyUnavailable — a dependency this route needs is down.
	ProblemDependencyUnavailable = corex.ProblemDependencyUnavailable
	// ProblemBadRequest — the request is malformed in a way none of the above
	// describes.
	ProblemBadRequest = corex.ProblemBadRequest
)

// Problem is an RFC 9457 problem details object. Aliased from corex.
type Problem = corex.Problem

// FieldError describes one field-level validation failure. Aliased from corex.
type FieldError = corex.FieldError

// NewProblem builds a Problem for a framework problem slug.
//
// A wrapper rather than a variable assignment: `var NewProblem =
// corex.NewProblem` would work, but it re-introduces exactly the mutability
// this extraction removed from ProblemTypeBase — anything in the process could
// reassign it.
func NewProblem(status int, slug, detail string) *Problem {
	return corex.NewProblem(status, slug, detail)
}

// WriteProblem is the one-line form: build and write in a single call.
//
//	rextension.WriteProblem(w, r, http.StatusUnauthorized,
//	    rextension.ProblemUnauthorized, "credentials were not accepted")
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, slug, detail string) {
	corex.WriteProblem(w, r, status, slug, detail)
}
