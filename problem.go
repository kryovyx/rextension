// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 Kryovyx

// Package rextension defines the minimal interface contract for Rex framework extensions.
//
// This file implements RFC 9457 "Problem Details for HTTP APIs" — the single
// error format for the whole stack (D30/O9).
//
// RFC 9457 obsoletes RFC 7807; the media type and member names are unchanged,
// so a client written against 7807 reads 9457 without modification.
//
// Before this, each extension invented its own shape. security wrote
// `http.Error(w, "401 Unauthorized: "+err.Error(), 401)` as text/plain,
// validation wrote a JSON `{status, message, errors}` envelope, ratelimit
// wrote plain text, health wrote plain text. A client talking to one
// application had to parse four formats and could not tell a rate limit from
// an authorization failure without string matching.
package rextension

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
	"net/http"
	"sort"
)

// ProblemMediaType is the media type RFC 9457 defines.
const ProblemMediaType = "application/problem+json"

// DefaultProblemTypeBase is the prefix for framework problem types.
//
// A URN rather than a URL. The RFC says `type` *should* be a URI that
// dereferences to human-readable documentation, but it does not require it —
// and a framework has no domain it can promise will still serve those pages.
// A URN is stable, identifies the problem unambiguously, and commits the
// project to no URL that could later 404.
//
// An application that does publish documentation overrides the base with
// ProblemTypeBase and gets dereferenceable types.
const DefaultProblemTypeBase = "urn:rex:problem:"

// DefaultInstanceBase is the prefix for the instance member.
const DefaultInstanceBase = "urn:rex:request:"

// Framework problem type slugs.
//
// Appended to the configured type base. These are the errors the framework and
// its extensions produce; an application adds its own.
const (
	// ProblemUnauthorized — authentication is required or failed.
	ProblemUnauthorized = "unauthorized"
	// ProblemForbidden — authenticated, but not permitted.
	ProblemForbidden = "forbidden"
	// ProblemNotFound — no route matches the request path.
	ProblemNotFound = "not-found"
	// ProblemMethodNotAllowed — the path exists under other methods.
	ProblemMethodNotAllowed = "method-not-allowed"
	// ProblemNotAcceptable — no representation matches the Accept header.
	ProblemNotAcceptable = "not-acceptable"
	// ProblemUnsupportedMediaType — the request body's Content-Type is not
	// supported.
	ProblemUnsupportedMediaType = "unsupported-media-type"
	// ProblemPayloadTooLarge — the request body exceeds the configured limit.
	ProblemPayloadTooLarge = "payload-too-large"
	// ProblemValidationFailed — the request body failed validation. Carries
	// the Errors extension member.
	ProblemValidationFailed = "validation-failed"
	// ProblemRateLimitExceeded — the caller has exceeded its rate limit.
	ProblemRateLimitExceeded = "rate-limit-exceeded"
	// ProblemInternal — an unexpected server-side failure.
	ProblemInternal = "internal"
	// ProblemDependencyUnavailable — a dependency this route needs is down.
	ProblemDependencyUnavailable = "dependency-unavailable"
	// ProblemBadRequest — the request is malformed in a way none of the above
	// describes.
	ProblemBadRequest = "bad-request"
)

// Problem is an RFC 9457 problem details object.
type Problem struct {
	// Type identifies the problem type. A URI reference; see
	// DefaultProblemTypeBase for why the default is a URN.
	Type string `json:"type"`

	// Title is a short, human-readable summary of the problem type. It must
	// not change from occurrence to occurrence — it describes the *type*, not
	// this instance.
	Title string `json:"title"`

	// Status is the HTTP status code, duplicated here so the body is
	// self-describing when it is logged or forwarded away from its response.
	Status int `json:"status"`

	// Detail is a human-readable explanation specific to this occurrence.
	//
	// SAFE TEXT ONLY. Never assign err.Error() here. An internal error's text
	// routinely carries a table name, a file path, a connection string, or the
	// shape of an internal service — none of which a client needs and all of
	// which help an attacker. The real cause belongs in the log, joined to
	// this response by Instance.
	Detail string `json:"detail,omitempty"`

	// Instance identifies this specific occurrence: the request identifier,
	// as DefaultInstanceBase + id.
	//
	// There is deliberately no separate trace_id member. `instance` is what
	// RFC 9457 provides for exactly this, and one identifier that appears in
	// both the response and the log line is what makes the Detail rule above
	// workable rather than merely restrictive — a client reports the value,
	// and the operator finds the real cause under it.
	Instance string `json:"instance,omitempty"`

	// Errors is an extension member carrying per-field validation failures.
	//
	// RFC 9457 §3.2 permits extension members; a consumer that does not know
	// this one ignores it. This is what replaces the validation extension's
	// bespoke envelope, so field-level detail survives the format change.
	Errors []FieldError `json:"errors,omitempty"`

	// Extra carries additional extension members, flattened into the object
	// on marshal. Reserved names are dropped rather than overwriting the
	// members above.
	Extra map[string]any `json:"-"`
}

// FieldError describes one field-level validation failure.
type FieldError struct {
	// Field is the name of the offending field, in the request body's own
	// terms — the JSON name, not the Go field name.
	Field string `json:"field"`
	// Rule is the constraint that failed, e.g. "required", "email", "max".
	Rule string `json:"rule,omitempty"`
	// Message is a human-readable explanation.
	Message string `json:"message"`
	// Value is the offending value, when echoing it is safe. Omitted for
	// anything that could be a credential.
	Value string `json:"value,omitempty"`
}

// reservedProblemMembers are the member names Extra must not overwrite.
var reservedProblemMembers = map[string]struct{}{
	"type": {}, "title": {}, "status": {}, "detail": {}, "instance": {}, "errors": {},
}

// MarshalJSONTo writes the problem document in one pass: the members RFC 9457
// declares, in the order it declares them, then the extension members.
//
// This replaced a MarshalJSON that marshalled the struct, unmarshalled the
// result into a map, merged Extra in, and marshalled the map. Two problems with
// that, one cosmetic and one not:
//
//   - Marshalling a map sorts its keys, so setting a single extension member
//     reordered the whole document — `type,title,status,detail` became
//     `detail,retry_after_seconds,status,title,type`. Object member order is
//     not significant to a parser, but a document whose shape depends on
//     whether an extension member happens to be present is harder to read in a
//     log and harder to pin in a test.
//   - It cost three passes over the document, on a path that only runs when
//     something has already gone wrong — including the 429, where it runs
//     under exactly the load the limit exists to shed.
//
// The method is MarshalJSONTo (encoding/json/v2) rather than MarshalJSON,
// because emitting members in a chosen order requires writing tokens rather
// than handing back a finished []byte. encoding/json v1 honours it too, so a
// caller on either API gets the same document.
func (p Problem) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	// Declared members, in RFC 9457 order. omitempty is applied by hand here:
	// the struct tags cannot reach a streaming encoder.
	if err := writeMember(enc, "type", p.Type); err != nil {
		return err
	}
	if err := writeMember(enc, "title", p.Title); err != nil {
		return err
	}
	if err := writeMember(enc, "status", p.Status); err != nil {
		return err
	}
	if p.Detail != "" {
		if err := writeMember(enc, "detail", p.Detail); err != nil {
			return err
		}
	}
	if p.Instance != "" {
		if err := writeMember(enc, "instance", p.Instance); err != nil {
			return err
		}
	}
	if len(p.Errors) > 0 {
		if err := writeMember(enc, "errors", p.Errors); err != nil {
			return err
		}
	}

	// Extension members last, in sorted order so the document is reproducible.
	// Map iteration order is randomised, and an error document that differs
	// between two identical failures cannot be diffed.
	if len(p.Extra) > 0 {
		keys := make([]string, 0, len(p.Extra))
		for k := range p.Extra {
			if _, reserved := reservedProblemMembers[k]; reserved {
				// Silently dropping is the lesser evil: letting an extension
				// member overwrite `status` would make the body contradict the
				// response it arrived in.
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := writeMember(enc, k, p.Extra[k]); err != nil {
				return fmt.Errorf("problem: extension member %q is not serialisable: %w", k, err)
			}
		}
	}

	return enc.WriteToken(jsontext.EndObject)
}

// writeMember writes one name/value pair to enc.
func writeMember(enc *jsontext.Encoder, name string, value any) error {
	if err := enc.WriteToken(jsontext.String(name)); err != nil {
		return err
	}
	return json.MarshalEncode(enc, value)
}

// ProblemTypeBase is the type prefix used by NewProblem.
//
// Package-level and mutable so an application can point problem types at its
// own documentation once, at startup, rather than threading a base through
// every construction site. Set it before serving; it is read on every
// NewProblem.
var ProblemTypeBase = DefaultProblemTypeBase

// InstanceBase is the instance prefix used by WithInstance.
var InstanceBase = DefaultInstanceBase

// NewProblem builds a Problem for a framework problem slug.
//
// Title defaults to the standard reason phrase for the status, which is
// exactly what a title should be: a description of the *type*, identical for
// every occurrence.
func NewProblem(status int, slug, detail string) *Problem {
	return &Problem{
		Type:   ProblemTypeBase + slug,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
}

// WithTitle overrides the title.
func (p *Problem) WithTitle(title string) *Problem {
	p.Title = title
	return p
}

// WithDetail sets the detail.
//
// SAFE TEXT ONLY — see Problem.Detail.
func (p *Problem) WithDetail(detail string) *Problem {
	p.Detail = detail
	return p
}

// WithInstance sets the instance member from a request identifier.
func (p *Problem) WithInstance(requestID string) *Problem {
	if requestID != "" {
		p.Instance = InstanceBase + requestID
	}
	return p
}

// WithErrors attaches field-level validation failures.
func (p *Problem) WithErrors(errs ...FieldError) *Problem {
	p.Errors = append(p.Errors, errs...)
	return p
}

// WithExtra attaches an extension member.
func (p *Problem) WithExtra(key string, value any) *Problem {
	if _, reserved := reservedProblemMembers[key]; reserved {
		return p
	}
	if p.Extra == nil {
		p.Extra = make(map[string]any, 1)
	}
	p.Extra[key] = value
	return p
}

// Write serialises the problem to w with the RFC 9457 media type and status.
//
// The media type is set **regardless of the request's Accept header** (O9).
// Negotiation applies to success responses: a client that asked for
// application/xml and then made a mistake is better served by a machine
// readable problem document it did not ask for than by a 406 carrying no
// information about what actually went wrong. RFC 9457 §3 anticipates this.
//
// Write is safe to call on a nil Problem, in which case it writes nothing.
func (p *Problem) Write(w http.ResponseWriter, r *http.Request) {
	if p == nil {
		return
	}
	if p.Status == 0 {
		p.Status = http.StatusInternalServerError
	}
	if p.Title == "" {
		p.Title = http.StatusText(p.Status)
	}

	body, err := json.Marshal(p)
	if err != nil {
		// Falling back to a hand-written document rather than to nothing: a
		// client must still receive a parseable problem, and the status is
		// the part that matters most.
		body = []byte(fmt.Sprintf(
			`{"type":%q,"title":%q,"status":%d}`,
			ProblemTypeBase+ProblemInternal, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError))
		p.Status = http.StatusInternalServerError
	}

	h := w.Header()
	h.Set("Content-Type", ProblemMediaType)
	// A problem response is specific to this request and must not be cached
	// by an intermediary as a representation of the resource.
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(p.Status)

	// HEAD carries no body, and writing one is a protocol error.
	if r != nil && r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(body)
}

// WriteProblem is the one-line form: build and write in a single call.
//
//	rextension.WriteProblem(w, r, http.StatusUnauthorized,
//	    rextension.ProblemUnauthorized, "credentials were not accepted")
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, slug, detail string) {
	NewProblem(status, slug, detail).Write(w, r)
}
