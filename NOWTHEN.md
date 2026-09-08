---
kind: library
forge: github
tracker: github-issues
test: make test
---

# arrest-go

A Go library: a DSL for generating OpenAPI 3.0/3.1 specifications, plus a Gin
integration that registers routes from the same definitions. It is consumed as a
Go module by other projects; there is nothing to deploy and no server anywhere.

## It is three modules, not one

There are three `go.mod` files — the root, `gin/`, and
`gin/examples/polymorphic/` — and `make test` runs all three. Anything that
touches dependencies must be applied across all three or CI fails on the ones
that were missed. This is the single most common way work here breaks.

Dependabot in particular opens a PR that bumps only one module, leaving the
others untidy and the checks red for a reason that has nothing to do with the
bump. `scripts/retidy-pr <branch>` rebases such a branch and re-tidies every
module; run it per Dependabot branch, and run it *before* judging any PR's check
status, or a mergeable PR will look broken. `scripts/retidy-prs` does the same
across every failing PR, including ones that are not Dependabot's, so it
force-pushes contributor branches — reach for it only deliberately.

## Releasing

There are no version tags and never have been, so consumers pick this up at a
pseudo-version from `master`. Merging to `master` is the whole of shipping
today. If a tagged release is ever wanted, that is a decision, not a routine.

## Upkeep

The recurring work here is dependency maintenance: the Dependabot sweep,
described in `.claude/skills/maintenance-deps`. CI (`.github/workflows/test.yaml`)
runs tests, race detection, coverage and `golangci-lint` on every push, for each
of the three modules.
