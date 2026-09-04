// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package rextension_test

import (
	jsonv1 "encoding/json"
	json "encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kryovyx/rextension"
)

// decodeProblem reads a recorded response as a generic JSON object, so a test
// can assert on the wire form rather than on the Go struct — which is what a
// client actually sees.
func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("the response body is not valid JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return got
}

func TestProblem_Write_shape(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	rextension.NewProblem(http.StatusUnauthorized, rextension.ProblemUnauthorized,
		"credentials were not accepted").
		WithInstance("req-123").
		Write(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != rextension.ProblemMediaType {
		t.Fatalf("Content-Type = %q, want %q", ct, rextension.ProblemMediaType)
	}

	got := decodeProblem(t, rec)
	want := map[string]any{
		"type":     "urn:rex:problem:unauthorized",
		"title":    "Unauthorized",
		"status":   float64(401),
		"detail":   "credentials were not accepted",
		"instance": "urn:rex:request:req-123",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}
}

// TestProblem_Write_ignores_Accept is the O9 decision, made explicit.
//
// Negotiation applies to success responses. A client that asked for XML and
// then made a mistake is better served by a machine-readable problem document
// it did not ask for than by a 406 carrying nothing about what went wrong.
func TestProblem_Write_ignores_Accept(t *testing.T) {
	for _, accept := range []string{"application/xml", "text/plain", "", "*/*", "application/json"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}

		rextension.WriteProblem(rec, req, http.StatusForbidden, rextension.ProblemForbidden, "no")

		if ct := rec.Header().Get("Content-Type"); ct != rextension.ProblemMediaType {
			t.Errorf("Accept: %q → Content-Type %q, want %q", accept, ct, rextension.ProblemMediaType)
		}
	}
}

// TestProblem_Write_omits_the_body_for_HEAD keeps the response protocol-legal.
func TestProblem_Write_omits_the_body_for_HEAD(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/x", nil)

	rextension.WriteProblem(rec, req, http.StatusNotFound, rextension.ProblemNotFound, "gone")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD must carry no body, got %d bytes", rec.Body.Len())
	}
	if ct := rec.Header().Get("Content-Type"); ct != rextension.ProblemMediaType {
		t.Errorf("Content-Type = %q, want %q", ct, rextension.ProblemMediaType)
	}
}

// TestProblem_omits_empty_optional_members keeps the document from carrying
// `"detail":""` and `"instance":""` on every error.
func TestProblem_omits_empty_optional_members(t *testing.T) {
	rec := httptest.NewRecorder()
	rextension.NewProblem(http.StatusInternalServerError, rextension.ProblemInternal, "").
		Write(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	got := decodeProblem(t, rec)
	for _, k := range []string{"detail", "instance", "errors"} {
		if _, present := got[k]; present {
			t.Errorf("%q should be omitted when empty, got %v", k, got[k])
		}
	}
	// The three required members are always present.
	for _, k := range []string{"type", "title", "status"} {
		if _, present := got[k]; !present {
			t.Errorf("%q must always be present", k)
		}
	}
}

// TestProblem_field_errors covers the extension member that replaces the
// validation extension's bespoke envelope.
func TestProblem_field_errors(t *testing.T) {
	rec := httptest.NewRecorder()
	rextension.NewProblem(http.StatusUnprocessableEntity, rextension.ProblemValidationFailed,
		"the request body failed validation").
		WithErrors(
			rextension.FieldError{Field: "email", Rule: "email", Message: "must be a valid email address"},
			rextension.FieldError{Field: "age", Rule: "min", Message: "must be at least 18", Value: "12"},
		).
		Write(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

	got := decodeProblem(t, rec)
	errs, ok := got["errors"].([]any)
	if !ok {
		t.Fatalf("errors is not an array: %v", got["errors"])
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 field errors, got %d", len(errs))
	}
	first := errs[0].(map[string]any)
	if first["field"] != "email" || first["rule"] != "email" {
		t.Errorf("first field error = %v", first)
	}
	// Value is omitted when unset, so a document does not echo empty strings.
	if _, present := first["value"]; present {
		t.Error("value should be omitted when unset")
	}
}

// TestProblem_extension_members covers Extra and the reserved-name guard.
func TestProblem_extension_members(t *testing.T) {
	rec := httptest.NewRecorder()
	rextension.NewProblem(http.StatusTooManyRequests, rextension.ProblemRateLimitExceeded, "slow down").
		WithExtra("retry_after_seconds", 30).
		WithExtra("limit", 100).
		// Reserved: must not overwrite the status the response actually
		// carries, or the body would contradict its own headers.
		WithExtra("status", 999).
		Write(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	got := decodeProblem(t, rec)
	if got["retry_after_seconds"] != float64(30) {
		t.Errorf("retry_after_seconds = %v, want 30", got["retry_after_seconds"])
	}
	if got["limit"] != float64(100) {
		t.Errorf("limit = %v, want 100", got["limit"])
	}
	if got["status"] != float64(http.StatusTooManyRequests) {
		t.Errorf("status = %v: a reserved member was overwritten", got["status"])
	}
}

// TestProblem_configurable_type_base covers the escape hatch for an
// application that publishes documentation and wants dereferenceable types.
func TestProblem_configurable_type_base(t *testing.T) {
	original := rextension.ProblemTypeBase
	t.Cleanup(func() { rextension.ProblemTypeBase = original })

	rextension.ProblemTypeBase = "https://api.example.com/problems/"

	rec := httptest.NewRecorder()
	rextension.WriteProblem(rec, httptest.NewRequest(http.MethodGet, "/x", nil),
		http.StatusNotFound, rextension.ProblemNotFound, "")

	got := decodeProblem(t, rec)
	if got["type"] != "https://api.example.com/problems/not-found" {
		t.Fatalf("type = %v", got["type"])
	}
}

// TestProblem_defaults_a_missing_status keeps a hand-built Problem from
// writing a status 0 response.
func TestProblem_defaults_a_missing_status(t *testing.T) {
	rec := httptest.NewRecorder()
	p := &rextension.Problem{Type: "urn:rex:problem:custom"}
	p.Write(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	got := decodeProblem(t, rec)
	if got["title"] != "Internal Server Error" {
		t.Fatalf("title = %v, want the reason phrase", got["title"])
	}
}

// TestProblem_nil_write_is_a_no_op lets a caller write a problem it may not
// have built.
func TestProblem_nil_write_is_a_no_op(t *testing.T) {
	rec := httptest.NewRecorder()
	var p *rextension.Problem
	p.Write(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Body.Len() != 0 {
		t.Fatalf("expected nothing written, got %q", rec.Body.String())
	}
}

// TestProblem_titles_describe_the_type_not_the_occurrence guards the property
// RFC 9457 requires of title: it must not vary between occurrences of the
// same problem type.
func TestProblem_titles_describe_the_type_not_the_occurrence(t *testing.T) {
	a := rextension.NewProblem(http.StatusForbidden, rextension.ProblemForbidden, "user 1 lacks role admin")
	b := rextension.NewProblem(http.StatusForbidden, rextension.ProblemForbidden, "user 2 lacks role editor")

	if a.Title != b.Title {
		t.Fatalf("titles differ between occurrences: %q vs %q", a.Title, b.Title)
	}
	if a.Detail == b.Detail {
		t.Fatal("detail is what varies per occurrence; these should differ")
	}
}

// TestProblem_every_framework_slug_is_distinct guards against a copy-paste
// duplicate in the slug list, which would make two different problems
// indistinguishable to a client branching on type.
func TestProblem_every_framework_slug_is_distinct(t *testing.T) {
	slugs := []string{
		rextension.ProblemUnauthorized,
		rextension.ProblemForbidden,
		rextension.ProblemNotFound,
		rextension.ProblemMethodNotAllowed,
		rextension.ProblemNotAcceptable,
		rextension.ProblemUnsupportedMediaType,
		rextension.ProblemPayloadTooLarge,
		rextension.ProblemValidationFailed,
		rextension.ProblemRateLimitExceeded,
		rextension.ProblemInternal,
		rextension.ProblemDependencyUnavailable,
		rextension.ProblemBadRequest,
	}
	seen := make(map[string]bool, len(slugs))
	for _, s := range slugs {
		if s == "" {
			t.Error("a framework problem slug is empty")
		}
		if strings.ContainsAny(s, " :/") {
			t.Errorf("slug %q contains a character that would break the URN", s)
		}
		if seen[s] {
			t.Errorf("duplicate problem slug %q", s)
		}
		seen[s] = true
	}
	if len(seen) != 12 {
		t.Fatalf("expected 12 distinct framework problem types, got %d", len(seen))
	}
}

// The builder methods are chained at almost every call site, so each must
// return the receiver — a method that returned a copy would silently drop
// everything set after it.
func TestProblem_builders_chain(t *testing.T) {
	p := rextension.NewProblem(http.StatusConflict, "conflict", "original detail").
		WithTitle("Custom Title").
		WithDetail("replaced detail")

	if p.Title != "Custom Title" {
		t.Fatalf("Title = %q", p.Title)
	}
	if p.Detail != "replaced detail" {
		t.Fatalf("Detail = %q", p.Detail)
	}
	if p.Status != http.StatusConflict {
		t.Fatalf("Status = %d", p.Status)
	}
}

func TestProblem_builders_return_receiver(t *testing.T) {
	p := rextension.NewProblem(http.StatusBadRequest, "bad-request", "d")
	if got := p.WithTitle("t"); got != p {
		t.Fatal("WithTitle returned a different pointer; chained calls would be lost")
	}
	if got := p.WithDetail("d2"); got != p {
		t.Fatal("WithDetail returned a different pointer; chained calls would be lost")
	}
}

// The document's member order must not depend on whether an extension member
// happens to be present.
//
// It did. The old MarshalJSON marshalled the struct, unmarshalled into a map,
// merged Extra, and marshalled the map — and marshalling a map sorts its keys,
// so one extension member reordered the whole document from RFC 9457 order into
// alphabetical order.
func TestProblem_member_order_is_stable_with_and_without_extras(t *testing.T) {
	base := rextension.NewProblem(http.StatusTooManyRequests, "rate-limit-exceeded", "slow down")
	withoutExtra, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}

	withExtra, err := json.Marshal(
		rextension.NewProblem(http.StatusTooManyRequests, "rate-limit-exceeded", "slow down").
			WithExtra("retry_after_seconds", 60))
	if err != nil {
		t.Fatal(err)
	}

	// The declared members keep RFC order in both, and the extension member is
	// appended rather than sorted into the middle.
	const declared = `{"type":"urn:rex:problem:rate-limit-exceeded","title":"Too Many Requests","status":429,"detail":"slow down"`
	for name, got := range map[string]string{"without extras": string(withoutExtra), "with extras": string(withExtra)} {
		if !strings.HasPrefix(got, declared) {
			t.Errorf("%s: declared members are not in RFC order:\n  got  %s\n  want prefix %s", name, got, declared)
		}
	}
	if !strings.HasSuffix(string(withExtra), `,"retry_after_seconds":60}`) {
		t.Errorf("the extension member is not appended after the declared ones: %s", withExtra)
	}
}

// Extension members are emitted in sorted order, so two identical failures
// produce byte-identical documents. Map iteration order is randomised.
func TestProblem_extension_members_are_deterministic(t *testing.T) {
	build := func() []byte {
		p := rextension.NewProblem(http.StatusTooManyRequests, "rate-limit-exceeded", "slow down").
			WithExtra("retry_after_seconds", 60).
			WithExtra("limit", 100).
			WithExtra("scope", "global")
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	first := build()
	for i := range 20 {
		if got := build(); string(got) != string(first) {
			t.Fatalf("run %d differs:\n  %s\n  %s", i, first, got)
		}
	}
	if !strings.Contains(string(first), `"limit":100,"retry_after_seconds":60,"scope":"global"`) {
		t.Errorf("extension members are not sorted: %s", first)
	}
}

// A reserved name in Extra is dropped rather than allowed to contradict the
// response it arrived in.
func TestProblem_extras_cannot_overwrite_declared_members(t *testing.T) {
	p := rextension.NewProblem(http.StatusConflict, "conflict", "d").
		WithExtra("status", 200).
		WithExtra("type", "urn:evil").
		WithExtra("safe", "kept")

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Contains(got, `"status":200`) || strings.Contains(got, "urn:evil") {
		t.Fatalf("a reserved member was overwritten: %s", got)
	}
	if !strings.Contains(got, `"status":409`) || !strings.Contains(got, `"safe":"kept"`) {
		t.Fatalf("unexpected document: %s", got)
	}
}

// A caller on encoding/json v1 must get the same document as one on v2 — the
// method is MarshalJSONTo, and v1 has to honour it.
func TestProblem_v1_and_v2_agree(t *testing.T) {
	p := rextension.NewProblem(http.StatusBadRequest, "validation-failed", "bad body").
		WithErrors(rextension.FieldError{Field: "email", Rule: "required", Message: "required"}).
		WithExtra("hint", "check the schema")

	v1, err := jsonv1.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(v1) != string(v2) {
		t.Fatalf("v1 and v2 disagree:\n  v1 %s\n  v2 %s", v1, v2)
	}
}

// Every optional member is emitted when set and omitted when not — the
// streaming encoder applies omitempty by hand, so each branch is a separate
// piece of code rather than a struct tag.
func TestProblem_optional_members(t *testing.T) {
	t.Run("all omitted when zero", func(t *testing.T) {
		p := &rextension.Problem{Type: "urn:t", Title: "T", Status: 400}
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		got := string(b)
		if got != `{"type":"urn:t","title":"T","status":400}` {
			t.Fatalf("unexpected document: %s", got)
		}
	})

	t.Run("all present when set", func(t *testing.T) {
		p := rextension.NewProblem(http.StatusUnprocessableEntity, "validation-failed", "bad body").
			WithInstance("req-123").
			WithErrors(rextension.FieldError{Field: "email", Rule: "required", Message: "required"})
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		got := string(b)
		for _, want := range []string{`"detail":"bad body"`, `"instance":`, `"errors":[`} {
			if !strings.Contains(got, want) {
				t.Errorf("missing %s in %s", want, got)
			}
		}
	})
}

// An extension member that cannot be serialised is reported with its name, so
// the offending member is identifiable rather than the whole document simply
// failing.
func TestProblem_unserialisable_extension_member(t *testing.T) {
	p := rextension.NewProblem(http.StatusConflict, "conflict", "d").
		WithExtra("bad", func() {}) // a func is not JSON

	_, err := json.Marshal(p)
	if err == nil {
		t.Fatal("expected an error for an unserialisable extension member")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Fatalf("the error does not name the member: %v", err)
	}
}

// Write emits the problem media type and the status from the document, so the
// body cannot contradict the response it arrived in.
func TestProblem_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rextension.NewProblem(http.StatusTeapot, "teapot", "short and stout").Write(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != rextension.ProblemMediaType {
		t.Fatalf("Content-Type = %q", ct)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if doc["status"] != float64(http.StatusTeapot) {
		t.Fatalf("status member = %v", doc["status"])
	}
}

// The reserved-name guard exists twice, and the second one matters.
//
// WithExtra refuses a reserved name, so the check inside the encoder is only
// reachable by assigning Extra directly — which is precisely the case it
// defends: a caller who builds the struct rather than using the builder must
// not be able to make `status` in the body disagree with the response.
func TestProblem_reserved_names_assigned_directly_are_dropped(t *testing.T) {
	p := &rextension.Problem{
		Type:   "urn:t",
		Title:  "T",
		Status: 409,
		Extra: map[string]any{
			"status": 200,
			"type":   "urn:evil",
			"kept":   true,
		},
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Contains(got, `"status":200`) || strings.Contains(got, "urn:evil") {
		t.Fatalf("a reserved member assigned directly reached the document: %s", got)
	}
	if !strings.Contains(got, `"kept":true`) {
		t.Fatalf("a non-reserved member was dropped: %s", got)
	}
}

// Write must still send a parseable problem when the document cannot be
// marshalled. Returning nothing would leave the client with a status and an
// empty body, which is the one outcome worse than a generic problem.
func TestProblem_Write_falls_back_when_marshalling_fails(t *testing.T) {
	p := &rextension.Problem{
		Type:   "urn:t",
		Title:  "T",
		Status: http.StatusConflict,
		Extra:  map[string]any{"bad": func() {}}, // a func is not JSON
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	p.Write(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 for a document that could not be built", rec.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("the fallback body is not parseable JSON: %v (%s)", err, rec.Body.String())
	}
	if doc["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("fallback status member = %v", doc["status"])
	}
}
