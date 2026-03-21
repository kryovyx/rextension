// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: (c) 2026 Kryovyx

package rextension_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/kryovyx/dix"
	"github.com/kryovyx/rextension"
	rxevent "github.com/kryovyx/rextension/event"
)

// ---- helpers ----

type tEvent struct {
	typ string
	ctx context.Context
}

func (e *tEvent) Type() string             { return e.typ }
func (e *tEvent) Context() context.Context { return e.ctx }

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

type tScheme struct{ n, t, d, c string }

func (s *tScheme) Name() string        { return s.n }
func (s *tScheme) Type() string        { return s.t }
func (s *tScheme) Description() string { return s.d }
func (s *tScheme) Challenge() string   { return s.c }

type tSecRoute struct{ s []string }

func (r *tSecRoute) RequiredSchemes() []string { return r.s }

type tRex struct {
	lg   rextension.Logger
	ct   dix.Container
	eb   rxevent.EventBus
	exts []rextension.Extension
	mws  []rextension.Middleware
	rts  []rextension.Route
}

func mkRex() *tRex {
	return &tRex{lg: mkLog(), ct: dix.New(), eb: mkBus()}
}
func (r *tRex) Logger() rextension.Logger                                 { return r.lg }
func (r *tRex) Container() dix.Container                                  { return r.ct }
func (r *tRex) EventBus() rxevent.EventBus                                { return r.eb }
func (r *tRex) Use(mw rextension.Middleware)                              { r.mws = append(r.mws, mw) }
func (r *tRex) RegisterRoute(rt rextension.Route) error                   { r.rts = append(r.rts, rt); return nil }
func (r *tRex) RegisterRouteToRouter(rt rextension.Route, n string) error { return nil }
func (r *tRex) CreateRouter(n string, c rextension.RouterConfig) error    { return nil }
func (r *tRex) WithExtensions(ext ...rextension.Extension)                { r.exts = append(r.exts, ext...) }

type tMinRex struct{}

func (r *tMinRex) Logger() rextension.Logger                            { return nil }
func (r *tMinRex) Container() dix.Container                             { return nil }
func (r *tMinRex) EventBus() rxevent.EventBus                           { return nil }
func (r *tMinRex) Use(rextension.Middleware)                            {}
func (r *tMinRex) RegisterRoute(rextension.Route) error                 { return nil }
func (r *tMinRex) RegisterRouteToRouter(rextension.Route, string) error { return nil }
func (r *tMinRex) CreateRouter(string, rextension.RouterConfig) error   { return nil }

// ---- Event ----

func TestEvent_TypeAndContext(t *testing.T) {
	ctx := context.Background()
	ev := &tEvent{typ: "test", ctx: ctx}
	var _ rxevent.Event = ev
	if ev.Type() != "test" {
		t.Errorf("Type=%q", ev.Type())
	}
	if ev.Context() != ctx {
		t.Error("ctx mismatch")
	}
}

func TestEvent_NilContext(t *testing.T) {
	ev := &tEvent{typ: "x"}
	if ev.Context() != nil {
		t.Error("expected nil")
	}
}

func TestBus_SubscribeEmit(t *testing.T) {
	b := mkBus()
	ok := false
	b.Subscribe("e", func(rxevent.Event) { ok = true })
	b.Emit(&tEvent{typ: "e", ctx: context.Background()})
	if !ok {
		t.Error("not called")
	}
}

func TestBus_MultiHandler(t *testing.T) {
	b := mkBus()
	n := 0
	b.Subscribe("e", func(rxevent.Event) { n++ })
	b.Subscribe("e", func(rxevent.Event) { n++ })
	b.Emit(&tEvent{typ: "e", ctx: context.Background()})
	if n != 2 {
		t.Errorf("n=%d", n)
	}
}

func TestBus_NoSub(t *testing.T) {
	b := mkBus()
	b.Emit(&tEvent{typ: "x", ctx: context.Background()})
	if len(b.emitted) != 1 {
		t.Error("emit count")
	}
}

func TestBus_SetLogClose(t *testing.T) {
	b := mkBus()
	b.SetLogger(mkLog())
	if b.log == nil {
		t.Error("log nil")
	}
	b.Close()
	if !b.closed {
		t.Error("not closed")
	}
}

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

// ---- LogLevel ----

func TestLogLevel_Values(t *testing.T) {
	vals := []struct {
		n string
		l rextension.LogLevel
		v int
	}{
		{"Trace", rextension.LogLevelTrace, 0},
		{"Debug", rextension.LogLevelDebug, 1},
		{"Info", rextension.LogLevelInfo, 2},
		{"Warn", rextension.LogLevelWarn, 3},
		{"Error", rextension.LogLevelError, 4},
		{"Off", rextension.LogLevelOff, 5},
	}
	for _, tt := range vals {
		t.Run(tt.n, func(t *testing.T) {
			if int(tt.l) != tt.v {
				t.Errorf("%d!=%d", int(tt.l), tt.v)
			}
		})
	}
}

func TestLogLevel_Order(t *testing.T) {
	if !(rextension.LogLevelTrace < rextension.LogLevelDebug &&
		rextension.LogLevelDebug < rextension.LogLevelInfo &&
		rextension.LogLevelInfo < rextension.LogLevelWarn &&
		rextension.LogLevelWarn < rextension.LogLevelError &&
		rextension.LogLevelError < rextension.LogLevelOff) {
		t.Error("bad order")
	}
}

// ---- Logger ----

func TestLog_AllMethods(t *testing.T) {
	l := mkLog()
	l.Info("i")
	l.Warn("w")
	l.Error("e")
	l.Debug("d")
	l.Trace("t")
	if len(l.is) != 1 || len(l.ws) != 1 || len(l.es) != 1 || len(l.ds) != 1 || len(l.ts) != 1 {
		t.Error("missing logs")
	}
}

func TestLog_SetLevel(t *testing.T) {
	l := mkLog()
	l.SetLogLevel(rextension.LogLevelDebug)
	if l.lvl != rextension.LogLevelDebug {
		t.Error("level")
	}
}

func TestLog_WithField(t *testing.T) {
	l := mkLog()
	n := l.WithField("k", "v").(*tLog)
	if n.fld["k"] != "v" {
		t.Error("field")
	}
}

func TestLog_WithFields(t *testing.T) {
	l := mkLog()
	n := l.WithFields(map[string]interface{}{"a": 1, "b": 2}).(*tLog)
	if n.fld["a"] != 1 || n.fld["b"] != 2 {
		t.Error("fields")
	}
}

func TestLog_WithError(t *testing.T) {
	l := mkLog()
	n := l.WithError(context.DeadlineExceeded).(*tLog)
	if n.errVal != context.DeadlineExceeded {
		t.Error("err")
	}
}

// ---- Middleware ----

func TestMW_Wrap(t *testing.T) {
	var mw rextension.Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-T", "1")
			next.ServeHTTP(w, r)
		})
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	rec := httptest.NewRecorder()
	mw(h).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("X-T") != "1" {
		t.Error("header")
	}
}

func TestMW_Chain(t *testing.T) {
	var o []string
	m1 := rextension.Middleware(func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { o = append(o, "1"); n.ServeHTTP(w, r) })
	})
	m2 := rextension.Middleware(func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { o = append(o, "2"); n.ServeHTTP(w, r) })
	})
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { o = append(o, "h") })
	m1(m2(h)).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if len(o) != 3 || o[0] != "1" || o[1] != "2" || o[2] != "h" {
		t.Errorf("order=%v", o)
	}
}

func TestMW_ShortCircuit(t *testing.T) {
	mw := rextension.Middleware(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(403) })
	})
	called := false
	h := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	rec := httptest.NewRecorder()
	mw(h).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if called {
		t.Error("called")
	}
	if rec.Code != 403 {
		t.Errorf("code=%d", rec.Code)
	}
}

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

// ---- Security ----

func TestSecRoute_WithSchemes(t *testing.T) {
	var _ rextension.SecuredRouteAccessor = &tSecRoute{}
	s := &tSecRoute{s: []string{"bearer", "apikey"}}
	if len(s.RequiredSchemes()) != 2 {
		t.Error("len")
	}
}

func TestSecRoute_Nil(t *testing.T) {
	s := &tSecRoute{}
	if s.RequiredSchemes() != nil {
		t.Error("expected nil")
	}
}

func TestSchemeAccessor_Fields(t *testing.T) {
	var _ rextension.SecuritySchemeAccessor = &tScheme{}
	s := &tScheme{n: "b", t: "http", d: "desc", c: "Bearer"}
	if s.Name() != "b" || s.Type() != "http" || s.Description() != "desc" || s.Challenge() != "Bearer" {
		t.Error("field")
	}
}

func TestGlobalSchemes_RegisterGet(t *testing.T) {
	rextension.RegisterSecuritySchemes(nil)
	if rextension.GetSecuritySchemes() != nil {
		t.Error("expect nil")
	}
	rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{
		&tScheme{n: "a"}, &tScheme{n: "b"},
	})
	r := rextension.GetSecuritySchemes()
	if len(r) != 2 || r[0].Name() != "a" || r[1].Name() != "b" {
		t.Error("mismatch")
	}
	rextension.RegisterSecuritySchemes(nil)
}

func TestGlobalSchemes_Snapshot(t *testing.T) {
	rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{&tScheme{n: "x"}})
	a := rextension.GetSecuritySchemes()
	b := rextension.GetSecuritySchemes()
	a[0] = &tScheme{n: "changed"}
	if b[0].Name() != "x" {
		t.Error("snapshot")
	}
	rextension.RegisterSecuritySchemes(nil)
}

func TestGlobalSchemes_Overwrite(t *testing.T) {
	rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{&tScheme{n: "old"}})
	rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{&tScheme{n: "n1"}, &tScheme{n: "n2"}})
	r := rextension.GetSecuritySchemes()
	if len(r) != 2 || r[0].Name() != "n1" {
		t.Error("overwrite")
	}
	rextension.RegisterSecuritySchemes(nil)
}

func TestGlobalSchemes_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			rextension.RegisterSecuritySchemes([]rextension.SecuritySchemeAccessor{&tScheme{n: "s"}})
		}()
		go func() { defer wg.Done(); _ = rextension.GetSecuritySchemes() }()
	}
	wg.Wait()
	rextension.RegisterSecuritySchemes(nil)
}

// ---- Rex ----

func TestDefaultRouterName(t *testing.T) {
	if rextension.DefaultRouterName != "default" {
		t.Error("name")
	}
}

func TestRouterCfg_Zero(t *testing.T) {
	c := rextension.RouterConfig{}
	if c.Addr != "" || c.BaseURL != "" || c.SSLVerify || c.ListenSSL || c.CertFile != nil || c.KeyFile != nil {
		t.Error("zero")
	}
}

func TestRouterCfg_Set(t *testing.T) {
	cf, kf := "c.pem", "k.pem"
	c := rextension.RouterConfig{Addr: ":9090", BaseURL: "/a", SSLVerify: true, ListenSSL: true, CertFile: &cf, KeyFile: &kf}
	if c.Addr != ":9090" || c.BaseURL != "/a" || !c.SSLVerify || !c.ListenSSL || *c.CertFile != cf || *c.KeyFile != kf {
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
