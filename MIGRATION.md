# rextension — migration to v0.3.0

**v0.2.1 → v0.3.0** · Go 1.27 · still pre-1.0 (**alpha**)

This is the `rextension` chapter of the REX v0.3.0 upgrade, written to stand alone:
you need nothing else to upgrade this module. If you use other modules of the
framework, each has its own guide — they are listed at the bottom.

---

## Before you start

- **Go 1.27 is a hard floor**, not a courtesy bump. The framework's own source
  uses `encoding/json/v2` and 1.27 struct-literal syntax. There is no build path
  on 1.26.
- **The modules depend on each other, so a partial upgrade does not compile.**
  Bump them in dependency order as a single change, using the sequence below,
  even if only one of them is the reason you are here.
- **Still alpha.** These are breaking changes and the versions stay pre-1.0
  deliberately. Pin exact versions; there is no compatibility promise yet.

---

## Upgrade in this order

```sh
# 1. the contract everything else compiles against
go get github.com/kryovyx/rextension@v0.3.0

# 2. the container, then the framework
go get github.com/kryovyx/dix@v0.2.0
go get github.com/kryovyx/rex@v0.3.0

# 3. the extensions you actually use
go get github.com/kryovyx/rextension-security@v0.6.0
go get github.com/kryovyx/rextension-validation@v0.3.0
go get github.com/kryovyx/rextension-openapi@v0.3.0
go get github.com/kryovyx/rextension-health@v0.3.0
go get github.com/kryovyx/rextension-metric@v0.3.0
go get github.com/kryovyx/rextension-swagger@v0.3.0

# 4. new, and optional
go get github.com/kryovyx/rextension-cors@v0.1.0
go get github.com/kryovyx/rextension-ratelimit@v0.1.0

# 5. the WebSocket side, if you want it. Additive: skipping this
#    changes nothing about the HTTP application.
go get github.com/kryovyx/rextension-wsx@v0.1.0

go mod tidy && go build ./...
```

`corex` is deliberately absent from that list. It is new in this release and
arrives as a dependency of `rextension`, which re-exports all of it as type
aliases; `go mod tidy` writes the require line for you. You name it explicitly
only if you call `corex.ConfigureProblems`.

> **If you use a `go.work` file**
>
> A workspace builds every module against sibling *source*, which hides exactly
> this kind of version skew — a module can be broken against its declared
> dependencies and still build for you. Verify with
> `GOWORK=off go build ./...` before you trust a green build.

---

## The one change behind this release

Rex now decides its routing and middleware once, at startup, instead of per
request. `New()` and the extension hooks **declare**; `Run()` **builds** — it
composes every middleware chain, builds the route tables, freezes them, and
only then binds listeners. Nothing is mutated after that.

Two consequences reach the whole framework:

- **Register routes and middleware from `OnInitialize` or `OnStart`, never from
  `OnReady`.** By `OnReady` the listeners are bound and the table is frozen, so
  registration now returns `ErrRouterFrozen` rather than mutating a trie that
  in-flight requests are reading.
- **Several things that used to fail silently now fail at startup** — an
  unknown security scheme name, a route requiring roles from a scheme that
  cannot enforce them, middleware registered for a router nothing creates, two
  routes collapsing to one trie key. Expect a boot failure or two on the first
  run. Each one is naming a bug that was already there.

The full account, with the new startup-time route validation that comes with
it, is in [`rex`'s guide](https://github.com/kryovyx/rex/blob/main/MIGRATION.md).

---

## What changed in `rextension`

The contract package. Every other module compiles against it, so upgrade this
first.

### It now depends on `corex`, and that is almost invisible

Roughly a third of what `rextension` declared is now declared in `corex`
instead: middleware and the priority scale, RFC 9457 `Problem`, `OriginPolicy`,
`Logger`, the event-bus contract, the DI contract, the `BodySchema` family, the
security-scheme accessors, `ClientIP`, and the reflection JSON Schema
generator. They moved because a second stack needs them — WSX is not HTTP and
must not import an HTTP contract, but it does need one problem format, one
origin allowlist and one schema generator rather than a second copy of each.

**Your imports do not change.** `rextension` re-exports every moved type as a
Go type alias:

```go
type Middleware = corex.Middleware
type Problem    = corex.Problem
```

An alias is the same type, not a convertible one, so a `corex.Middleware`
satisfies a parameter declared as `rextension.Middleware` and the other way
round, with no conversion at any call site. Constants are re-declared by value
and functions are thin wrappers, which have the same effect. This is checked
rather than asserted: `corex_alias_test.go` in the `rextension` repository
holds twenty-nine assignments in both directions, and it is a compile error if
any of them ever stops being an identity.

Two consequences you can see:

- `go mod tidy` adds `github.com/kryovyx/corex` to your `go.mod` as an indirect
  dependency. That is expected.
- `OriginPolicyError`'s message now begins `corex: ` rather than `rextension: `,
  because `corex` is the module that declares it. Nothing in the framework
  branches on the text — `rextension-cors` matches the type with `errors.As` —
  but if *your* code matches that string, match the type instead.

### `ProblemTypeBase` / `InstanceBase` — removed

These were mutable package-level variables. They are gone, and nothing in
`rextension` replaces them.

```go
// before
rextension.ProblemTypeBase = "https://api.example.com/problems/"
rextension.InstanceBase    = "https://api.example.com/traces/"

// after
corex.ConfigureProblems(
    corex.WithProblemTypeBase("https://api.example.com/problems/"),
    corex.WithInstanceBase("https://api.example.com/traces/"),
)
```

**Why this one breaks rather than shimming.** A variable has no alias form, so
the only shim available is `var ProblemTypeBase = corex.ProblemTypeBase` — and
that makes a *copy*. Setting `rextension.ProblemTypeBase` would then configure
nothing at all, silently, while the framework kept reading its own value. A
compile error is better than a setting that stops working without saying so.

They were also process-wide mutable configuration read on every problem
construction, which is the shape D21 removed from the security scheme registry
in this same release. The move was a free opportunity to close an
already-diagnosed defect rather than carry it across.

This is the only breaking symbol in the whole `corex` extraction.

### `RegisterSecuritySchemes` / `GetSecuritySchemes` — removed

The package-level scheme registry is gone. It was process-global state that two
Rex instances in one binary would fight over. The security extension now
publishes a `SchemeRegistry` into the DI container, and OpenAPI resolves it
through the `rextension.SchemeRegistry` interface — same decoupling, no global.

If you called either function directly, resolve the registry from the container
instead.

### The DI contract moved to `rextension/di` — import change

`Resolver`, `Container` and `Scope` now live in a leaf package, `rextension/di`,
because `rextension` imports `rextension/route` and `route.Context` needs a
resolver — an import cycle otherwise. The old names remain as aliases, so most
code is unaffected:

```go
type Resolver  = di.Resolver
type Container = di.Container
type Scope     = di.Scope
```

What does change: `Rex.Container()` returns `rextension.Container`, not
`dix.Container`. A handler or extension that named the concrete `dix` type needs
the interface instead. This is what lets an extension depend on `rextension`
alone, with no `dix` in its `go.mod`.

### `RouterConfig.SSLVerify` — removed

It configured nothing at all — the value was read by no code on any path. Delete
it from your config; a field that suggests TLS verification is happening when
none is, is worse than its absence.

The same struct gains real controls that were previously unbounded. A server
with no timeouts holds a connection open indefinitely, and a request body with
no cap is a memory limit set by whoever is calling you:

```go
ReadHeaderTimeout time.Duration
ReadTimeout       time.Duration
WriteTimeout      time.Duration
IdleTimeout       time.Duration
MaxHeaderBytes    int
MaxBodyBytes      int64
```

The `default:` struct tags are also gone. They were decorative — nothing read
them. Build config with `NewDefaultConfig()` and the `With*` options.

`MaxHeaderValueCount` is new alongside them, capping the *number* of header
values rather than their total size — a few thousand tiny headers stay well
inside 1 MiB while still forcing the server to allocate and hash every one.
`net/http` applies its own default of 500 whether or not you set it, so leaving
it zero is already safe; the field exists so the limit is adjustable alongside
the other five instead of being the one bound the configuration cannot reach.

### `Problem.MarshalJSON` is now `MarshalJSONTo`

⚠ **Breaking if you call it directly**, which is unusual — `json.Marshal(p)` and
`p.Write(w, r)` are unaffected, and `encoding/json` v1 honours the new method,
so a consumer on either JSON API gets the same document.

The output changed, for the better. The old implementation marshalled the
struct, unmarshalled the result into a map, merged `Extra`, and marshalled the
map — and marshalling a map sorts its keys, so **one extension member reordered
the whole document**:

```jsonc
// before, without an extension member — RFC 9457 order
{"type":"…","title":"Too Many Requests","status":429,"detail":"slow down"}

// before, with one — alphabetical, because a map was marshalled
{"detail":"slow down","retry_after_seconds":60,"status":429,"title":"…","type":"…"}

// now — declared members in RFC order, extension members appended, sorted
{"type":"…","title":"Too Many Requests","status":429,"detail":"slow down","retry_after_seconds":60}
```

Object member order is not significant to a parser, so this breaks no client
that parses properly. It does change golden-file tests and log greps. It also
removes two JSON round-trips from every error response — including the 429,
which ran them under exactly the load the limit exists to shed.

### Two event constructors changed signature

Only affects you if you construct these events yourself, which is unusual.

```go
NewRouterRouteRegisteredEvent(ctx, routerName, rt, baseURL)
NewRouterRequestHandledEvent(ctx, routerName, req, rw, dur,
    status, bytesWritten, routePattern)
```

The request-handled event carries real values now. Status and byte count come
from a response-writer wrapper rather than being assumed, and `routePattern` is
the registered pattern — which is what makes metrics labels bounded instead of
one series per distinct URL.

### `Source()` now returns the router name for every router event

`NewRouterInitializedEvent` never set the event's `source`, alone among the six
constructors, so `Source()` returned `""` for it and the router name for every
other one. Three levels of struct nesting is what hid it.

Latent rather than live: `Source()` is not on the `Event` interface, so only
code holding a concrete event type could observe it. Still a behaviour change if
you do.

### New, and worth knowing about

- `Problem`, `NewProblem`, `WriteProblem`, `FieldError` — RFC 9457
  `application/problem+json`, now the error shape across every extension.
- `OriginPolicy` — one origin allowlist, shared by CORS and CSRF, so the two
  cannot drift apart.
- `RouteInfo`, `RouteValidator`, `PerRouteMiddleware`, `PerRouterMiddleware`.
- `BodySchema` with `Scalar` / `OneOf` / `AnyOf` / `AllOf`, plus
  `BodySchemaProvider` — moved here from validation so OpenAPI can read schemas
  without importing it.
- `BodyLimitedRoute` — a route can declare its own body cap.
- Sentinels: `ErrRouterExists`, `ErrRouterUnknown`, `ErrRouterFrozen`.

---

## Verification

- [ ] `GOWORK=off go build ./...` passes — a workspace hides version skew, so
      this is the check that matters.
- [ ] `go test -race ./...` passes. The container and event bus were both
      unsynchronised before; if your tests never ran with `-race`, run them now.
- [ ] The application **starts**. Startup failures are the point of this
      release; each one names a pre-existing bug.

---

*Part of the REX v0.3.0 upgrade. The other guides:*

- [`rex`](https://github.com/kryovyx/rex/blob/main/MIGRATION.md) — v0.2.1 → **v0.3.0**
- [`dix`](https://github.com/kryovyx/dix/blob/main/MIGRATION.md) — v0.1.0 → **v0.2.0**
- [`rextension-security`](https://github.com/kryovyx/rextension-security/blob/main/MIGRATION.md) — v0.5.0 → **v0.6.0**
- [`rextension-validation`](https://github.com/kryovyx/rextension-validation/blob/main/MIGRATION.md) — v0.2.0 → **v0.3.0**
- [`rextension-openapi`](https://github.com/kryovyx/rextension-openapi/blob/main/MIGRATION.md) — v0.2.1 → **v0.3.0**
- [`rextension-health`](https://github.com/kryovyx/rextension-health/blob/main/MIGRATION.md) — v0.2.1 → **v0.3.0**
- [`rextension-metric`](https://github.com/kryovyx/rextension-metric/blob/main/MIGRATION.md) — v0.2.1 → **v0.3.0**
- [`rextension-swagger`](https://github.com/kryovyx/rextension-swagger/blob/main/MIGRATION.md) — v0.2.1 → **v0.3.0**
- [`corex`](https://github.com/kryovyx/corex/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`rextension-cors`](https://github.com/kryovyx/rextension-cors/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`rextension-ratelimit`](https://github.com/kryovyx/rextension-ratelimit/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`wsxtension`](https://github.com/kryovyx/wsxtension/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`wsx`](https://github.com/kryovyx/wsx/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`wsxtension-asyncapi`](https://github.com/kryovyx/wsxtension-asyncapi/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`wsxtension-lens`](https://github.com/kryovyx/wsxtension-lens/blob/main/MIGRATION.md) — new in this release, **v0.1.0**
- [`rextension-wsx`](https://github.com/kryovyx/rextension-wsx/blob/main/MIGRATION.md) — new in this release, **v0.1.0**

*Every one of them stands alone; read only the ones for modules you use.*
