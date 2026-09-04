// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package rextension_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kryovyx/rextension"
	rxevent "github.com/kryovyx/rextension/event"
)

// ---- helpers ----

type tBus struct {
	subs    map[string][]rxevent.EventHandler
	emitted []rxevent.Event
	log     rxevent.BusLogger
	closed  bool
}

func mkBus() *tBus { return &tBus{subs: make(map[string][]rxevent.EventHandler)} }

func (b *tBus) Subscribe(t string, h rxevent.EventHandler) {
	b.subs[t] = append(b.subs[t], h)
}
func (b *tBus) Emit(e rxevent.Event) {
	b.emitted = append(b.emitted, e)
	for _, h := range b.subs[e.Type()] {
		h(e)
	}
}
func (b *tBus) SetLogger(l rxevent.BusLogger) { b.log = l }
func (b *tBus) Close()                        { b.closed = true }

type tLog struct {
	is, ws, es, ds, ts []string
	lvl                rextension.LogLevel
	fld                map[string]interface{}
	errVal             error
}

func mkLog() *tLog { return &tLog{fld: map[string]interface{}{}} }

func (l *tLog) Info(f string, a ...interface{})   { l.is = append(l.is, f) }
func (l *tLog) Warn(f string, a ...interface{})   { l.ws = append(l.ws, f) }
func (l *tLog) Error(f string, a ...interface{})  { l.es = append(l.es, f) }
func (l *tLog) Debug(f string, a ...interface{})  { l.ds = append(l.ds, f) }
func (l *tLog) Trace(f string, a ...interface{})  { l.ts = append(l.ts, f) }
func (l *tLog) SetLogLevel(v rextension.LogLevel) { l.lvl = v }
func (l *tLog) WithField(k string, v interface{}) rextension.Logger {
	n := mkLog()
	for xk, xv := range l.fld {
		n.fld[xk] = xv
	}
	n.fld[k] = v
	return n
}
func (l *tLog) WithFields(m map[string]interface{}) rextension.Logger {
	n := mkLog()
	for xk, xv := range l.fld {
		n.fld[xk] = xv
	}
	for xk, xv := range m {
		n.fld[xk] = xv
	}
	return n
}
func (l *tLog) WithError(e error) rextension.Logger {
	n := mkLog()
	n.errVal = e
	return n
}

type tExt struct {
	init, start, ready, stop, shut      bool
	initE, startE, readyE, stopE, shutE error
}

func (e *tExt) OnInitialize(_ context.Context, _ rextension.Rex) error { e.init = true; return e.initE }
func (e *tExt) OnStart(_ context.Context, _ rextension.Rex) error      { e.start = true; return e.startE }
func (e *tExt) OnReady(_ context.Context, _ rextension.Rex) error      { e.ready = true; return e.readyE }
func (e *tExt) OnStop(_ context.Context, _ rextension.Rex) error       { e.stop = true; return e.stopE }
func (e *tExt) OnShutdown(_ context.Context, _ rextension.Rex) error   { e.shut = true; return e.shutE }

type tRoute struct{ m, p string }

func (r *tRoute) Method() string { return r.m }
func (r *tRoute) Path() string   { return r.p }

// tContainer is a do-nothing rextension.Container.
//
// The test used to construct a real dix.New(), which is what kept `dix` in
// this module's go.mod even though no non-test file references it. The
// compile-time proof that dix.Container satisfies rextension.Container lives
// in the rex module, which imports both (D23).
type tContainer struct{}

func (c *tContainer) Resolve(any) error        { return nil }
func (c *tContainer) ResolveAll(any) error     { return nil }
func (c *tContainer) Singleton(any) error      { return nil }
func (c *tContainer) Scoped(any) error         { return nil }
func (c *tContainer) Transient(any) error      { return nil }
func (c *tContainer) Instance(any) error       { return nil }
func (c *tContainer) Unbind(any) (bool, error) { return false, nil }

var _ rextension.Container = (*tContainer)(nil)

type tRex struct {
	lg        rextension.Logger
	ct        rextension.Container
	eb        rxevent.EventBus
	exts      []rextension.Extension
	mws       []rextension.Middleware
	rts       []rextension.Route
	routerMws []routerMW
	perRoute  []perRouteMW
	perRouter []perRouterMW
}

// routerMW records a UseOnRouter call.
type routerMW struct {
	name     string
	mw       rextension.Middleware
	priority int
}

// perRouteMW records a UsePerRoute call.
type perRouteMW struct {
	f        rextension.PerRouteMiddleware
	priority int
}

// perRouterMW records a UsePerRouter call.
type perRouterMW struct {
	f        rextension.PerRouterMiddleware
	priority int
}

func mkRex() *tRex {
	return &tRex{lg: mkLog(), ct: &tContainer{}, eb: mkBus()}
}
func (r *tRex) Logger() rextension.Logger       { return r.lg }
func (r *tRex) Container() rextension.Container { return r.ct }
func (r *tRex) EventBus() rxevent.EventBus      { return r.eb }
func (r *tRex) Use(mw rextension.Middleware)    { r.mws = append(r.mws, mw) }

// UseOnRouter and UsePerRoute record their arguments so tests can assert what
// an extension attached, and to which routes.
func (r *tRex) UseOnRouter(name string, mw rextension.Middleware, priority int) {
	r.routerMws = append(r.routerMws, routerMW{name: name, mw: mw, priority: priority})
}

func (r *tRex) UsePerRoute(f rextension.PerRouteMiddleware, priority int) {
	r.perRoute = append(r.perRoute, perRouteMW{f: f, priority: priority})
}

func (r *tRex) UsePerRouter(f rextension.PerRouterMiddleware, priority int) {
	r.perRouter = append(r.perRouter, perRouterMW{f: f, priority: priority})
}

func (r *tRex) RegisterRoute(rt rextension.Route) error                   { r.rts = append(r.rts, rt); return nil }
func (r *tRex) RegisterRouteToRouter(rt rextension.Route, n string) error { return nil }
func (r *tRex) CreateRouter(n string, c rextension.RouterConfig) error    { return nil }
func (r *tRex) WithExtensions(ext ...rextension.Extension)                { r.exts = append(r.exts, ext...) }

type tMinRex struct{}

func (r *tMinRex) Logger() rextension.Logger                            { return nil }
func (r *tMinRex) Container() rextension.Container                      { return nil }
func (r *tMinRex) EventBus() rxevent.EventBus                           { return nil }
func (r *tMinRex) Use(rextension.Middleware)                            {}
func (r *tMinRex) UseOnRouter(string, rextension.Middleware, int)       {}
func (r *tMinRex) UsePerRoute(rextension.PerRouteMiddleware, int)       {}
func (r *tMinRex) UsePerRouter(rextension.PerRouterMiddleware, int)     {}
func (r *tMinRex) RegisterRoute(rextension.Route) error                 { return nil }
func (r *tMinRex) RegisterRouteToRouter(rextension.Route, string) error { return nil }
func (r *tMinRex) CreateRouter(string, rextension.RouterConfig) error   { return nil }

// The Event and EventBus contract tests moved to corex/event with the
// declarations (W22). The bus is still reached through rextension/event, and
// that package's own suite covers the re-exports.

// ---- Extension ----

func TestExt_Lifecycle(t *testing.T) {
	e := &tExt{}
	ctx := context.Background()
	e.OnInitialize(ctx, nil)
	e.OnStart(ctx, nil)
	e.OnReady(ctx, nil)
	e.OnStop(ctx, nil)
	e.OnShutdown(ctx, nil)
	if !e.init || !e.start || !e.ready || !e.stop || !e.shut {
		t.Error("lifecycle incomplete")
	}
}

func TestExt_Error(t *testing.T) {
	e := &tExt{initE: context.DeadlineExceeded}
	if err := e.OnInitialize(context.Background(), nil); err != context.DeadlineExceeded {
		t.Errorf("err=%v", err)
	}
}

// The LogLevel, Logger and Middleware tests moved to corex with the
// declarations. What has to be tested *here* is that the aliases resolve to
// those declarations rather than to copies — see corex_alias_test.go.

// ---- Route ----

func TestRoute_MethodPath(t *testing.T) {
	var _ rextension.Route = &tRoute{}
	r := &tRoute{m: "GET", p: "/api"}
	if r.Method() != "GET" || r.Path() != "/api" {
		t.Error("mismatch")
	}
}

func TestRoute_AllHTTPMethods(t *testing.T) {
	for _, m := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
		r := &tRoute{m: m}
		if r.Method() != m {
			t.Errorf("%s", m)
		}
	}
}

// The security accessor and SchemeRegistry contract tests moved to corex,
// along with the note about the four TestGlobalSchemes_* tests that D21
// deleted with the package-level registry they exercised.

// ---- Rex ----

func TestDefaultRouterName(t *testing.T) {
	if rextension.DefaultRouterName != "default" {
		t.Error("name")
	}
}

func TestRouterCfg_Zero(t *testing.T) {
	c := rextension.RouterConfig{}
	if c.Addr != "" || c.BaseURL != "" || c.ListenSSL || c.CertFile != nil || c.KeyFile != nil || c.TLSConfig != nil {
		t.Error("zero")
	}
}

// TestRouterCfg_TLSConfig_defaults_to_nil pins the backwards compatibility the new
// field depends on: a config that never mentions TLSConfig must be indistinguishable
// from one written before the field existed, so the listener keeps using
// CertFile/KeyFile. Only an explicitly set TLSConfig takes precedence.
func TestRouterCfg_TLSConfig_defaults_to_nil(t *testing.T) {
	cf, kf := "c.pem", "k.pem"
	c := rextension.RouterConfig{Addr: ":9090", BaseURL: "/a", ListenSSL: true, CertFile: &cf, KeyFile: &kf}
	if c.TLSConfig != nil {
		t.Error("TLSConfig must stay nil unless set explicitly")
	}
}

func TestRouterCfg_TLSConfig_set(t *testing.T) {
	tc := &tls.Config{MinVersion: tls.VersionTLS12}
	c := rextension.RouterConfig{Addr: ":9090", TLSConfig: tc}
	if c.TLSConfig != tc {
		t.Error("TLSConfig must be stored verbatim")
	}
	if c.CertFile != nil || c.KeyFile != nil {
		t.Error("TLSConfig must not imply cert paths")
	}
}

func TestRouterCfg_Set(t *testing.T) {
	cf, kf := "c.pem", "k.pem"
	c := rextension.RouterConfig{Addr: ":9090", BaseURL: "/a", ListenSSL: true, CertFile: &cf, KeyFile: &kf}
	if c.Addr != ":9090" || c.BaseURL != "/a" || !c.ListenSSL || *c.CertFile != cf || *c.KeyFile != kf {
		t.Error("set")
	}
}

func TestRex_Iface(t *testing.T) {
	r := mkRex()
	var _ rextension.Rex = r
	if r.Logger() == nil || r.Container() == nil || r.EventBus() == nil {
		t.Error("nil")
	}
}

func TestRex_UseMW(t *testing.T) {
	r := mkRex()
	r.Use(func(n http.Handler) http.Handler { return n })
	if len(r.mws) != 1 {
		t.Error("mw")
	}
}

func TestRex_RegRoute(t *testing.T) {
	r := mkRex()
	r.RegisterRoute(&tRoute{m: "GET", p: "/t"})
	if len(r.rts) != 1 {
		t.Error("route")
	}
}

func TestWithExt_Valid(t *testing.T) {
	r := mkRex()
	rextension.WithExtension(&tExt{})(r)
	if len(r.exts) != 1 {
		t.Errorf("len=%d", len(r.exts))
	}
}

func TestWithExt_NoMethod(t *testing.T) {
	r := &tMinRex{}
	rextension.WithExtension(&tExt{})(r) // no panic
}

func TestWithExts_Multi(t *testing.T) {
	r := mkRex()
	rextension.WithExtensions(&tExt{}, &tExt{})(r)
	if len(r.exts) != 2 {
		t.Errorf("len=%d", len(r.exts))
	}
}

func TestWithExts_Empty(t *testing.T) {
	r := mkRex()
	rextension.WithExtensions()(r)
	if len(r.exts) != 0 {
		t.Error("empty")
	}
}

func TestRex_LifecycleInteg(t *testing.T) {
	r := mkRex()
	e := &tExt{}
	e.OnInitialize(context.Background(), r)
	e.OnStart(context.Background(), r)
	e.OnReady(context.Background(), r)
	e.OnStop(context.Background(), r)
	e.OnShutdown(context.Background(), r)
	if !e.init || !e.start || !e.ready || !e.stop || !e.shut {
		t.Error("lifecycle")
	}
}
