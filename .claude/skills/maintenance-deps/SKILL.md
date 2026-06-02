---
name: maintenance-deps
description: Run dependency maintenance for this project — re-tidies go.mod across all modules for any failing Dependabot PRs (via scripts/retidy-pr), then executes the zed:dependabot-sweep plugin skill.
---

# Dependency Maintenance

Run the full Dependabot maintenance sweep for this repository.

This is a multi-module Go project — it currently has multiple `go.mod` files
(the root module, `gin/`, and `gin/examples/polymorphic/`). When Dependabot bumps
a dependency it often updates only one module's `go.mod`/`go.sum`, leaving the
modules out of sync and the CI tests failing. The `scripts/retidy-pr` helper fixes
this by rebasing a branch onto its base and running `go mod tidy` across every
module. Because the sweep relies on PR check status to decide what is mergeable,
this re-tidy **must** happen first — otherwise the sweep sees spurious failures and
either skips mergeable PRs or churns on them. This is a known brittle spot for this
project, so the pre-check below is mandatory.

## Project-specific

Run this **before** invoking `zed:dependabot-sweep`:

1. Find the open Dependabot PRs that currently have failing checks. Match the
   conclusions that `scripts/retidy-prs` treats as failures (`FAILURE`,
   `CANCELLED`, `TIMED_OUT`):

   ```bash
   gh pr list --author "app/dependabot" --state open \
     --json number,headRefName,statusCheckRollup \
     --jq '.[] | select([.statusCheckRollup[]? | select(.conclusion=="FAILURE" or .conclusion=="CANCELLED" or .conclusion=="TIMED_OUT")] | length > 0) | "\(.number) \(.headRefName)"'
   ```

2. For **each** Dependabot branch listed above, run `scripts/retidy-pr` on that
   specific branch to synchronize `go.mod` / `go.sum` across all modules:

   ```bash
   ./scripts/retidy-pr <headRefName>
   ```

   `retidy-pr` creates a worktree for the branch, rebases it onto its base
   (preferring the PR's changes on conflicting `go.mod` lines), runs `go mod tidy`
   in every module directory, commits any changes, and force-pushes with lease.
   Run it once per failing Dependabot branch and let each invocation complete.

   > **Why not `scripts/retidy-prs`?** That helper processes **all** open PRs with
   > failing test/lint/CI checks — not just Dependabot's — and rebases and
   > force-pushes each one. Running it here could rewrite unrelated contributor
   > branches. Drive `retidy-pr` per Dependabot branch instead. Only fall back to
   > `./scripts/retidy-prs` if you have confirmed every failing PR is Dependabot's
   > and intend to re-tidy all of them.

3. **Verify the re-tidy actually succeeded** before proceeding. Confirm each
   `retidy-pr` invocation exited 0 and did not report any `[ERROR]` lines. Then
   re-run the `gh pr list` command from step 1 and confirm the previously-failing
   Dependabot PRs either now have passing/pending checks or are failing for a
   reason unrelated to go.mod sync. Do **not** continue to the sweep while a PR is
   still failing purely because its modules are out of tidy — investigate and
   re-run `./scripts/retidy-pr <branch>` for that branch first.

   If a re-tidy genuinely cannot be completed for some PR (e.g. an unresolvable
   rebase conflict), note which PR and why, and surface it in the final report
   rather than silently proceeding.

Only once the failing Dependabot PRs have been re-tidied and verified should you
continue to the standard sweep below.

## Steps

1. Complete the **Project-specific** re-tidy verification above first.
2. Invoke the `zed:dependabot-sweep` skill and let it run to completion. It will
   unblock stuck Dependabot PRs, merge ready PRs, fix vulnerability alerts, update
   the changelog, and open a PR as appropriate for this repository.
3. Report what the re-tidy pre-check and the sweep did, including any PRs that
   could not be re-tidied.
