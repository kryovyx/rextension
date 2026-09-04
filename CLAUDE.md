# rextension — the extension contract

`github.com/kryovyx/rextension`. Part of the REX framework: developed in a `go.work` workspace
alongside its siblings, released as its own module.

## Boundaries

The contract module: **interfaces and types only**, importing **stdlib only**.
No sibling module, no third-party package.

Its import list *is* the dependency graph of every extension that exists,
including third-party ones. Adding a dependency here adds it to all of them, so
anything concrete belongs in `rex` or in an extension instead.

## Working here

- **Never `go build`.** Syntax-check with `go vet ./...`.
- **`go test -race ./...` always.**
- **Tests are per branch, not per coverage number.** Every branch of every
  function gets its own case; the README's coverage figure is recomputed from
  a measurement, never hand-edited.
- **No `replace` directives** in `go.mod`.
- **Commits:** `<gitmoji><type>(<scope>): <description>` — feat, fix, docs,
  style, refactor, test, chore.
- `make check` here runs fmt, vet and race tests for this module alone.
- Default branch is `main`. **Never push without asking** — github
  authenticates with a hardware key that needs a physical tap, so an
  unattended push hangs and then fails.

Design decisions are numbered (D…/O…/W…) and recorded in the workspace this
module is developed in, not in this repo. If a rule here looks arbitrary, it is
load-bearing — ask before removing it.
