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

Releases are tagged. A `release/vX.Y.Z` branch is the dry run
(`.github/workflows/prepare.yaml` checks `version.txt`, `gin/go.mod`, and the
`Changes.md` heading); merging it and pushing the `vX.Y.Z` tag runs
`.github/workflows/release.yaml`, which repeats the checks, pushes the
`gin/vX.Y.Z` tag the nested Gin module needs, and publishes the GitHub release
from the changelog section. The `/release` skill in `.claude/skills/release`
drives this and asks a human to confirm the version, so cutting a release is
not unattended work.

Two checks are strict enough to be worth knowing in advance: `gin/go.mod` must
`require github.com/zostay/arrest-go vX.Y.Z` (its `replace` hides that line
locally, but consumers see it), and the first line of `Changes.md` must be
exactly `## X.Y.Z  YYYY-MM-DD` (two spaces) with the date being *today in
America/Chicago* at the moment the workflow runs — so a release straddling the
Central midnight fails at the tag even though the branch passed.

Between releases, changes land under a `## Unreleased` heading at the top of
`Changes.md`; every PR that changes behaviour should add a bullet there.

## Upkeep

The recurring work here is dependency maintenance: the Dependabot sweep,
described in `.claude/skills/maintenance-deps`. CI (`.github/workflows/test.yaml`)
runs tests, race detection, coverage and `golangci-lint` on every push, for each
of the three modules.
