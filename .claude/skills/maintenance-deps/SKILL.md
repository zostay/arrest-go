---
name: maintenance-deps
description: Run dependency maintenance for this project — verifies retidy-prs has synced go.mod across failing Dependabot PRs, then executes the zed:dependabot-sweep plugin skill.
---

# Dependency Maintenance

Run the full Dependabot maintenance sweep for this repository.

This is a multi-module Go project (root module plus the `gin/` submodule). When
Dependabot bumps a dependency it often updates only one module's `go.mod`/`go.sum`,
leaving the modules out of sync and the CI tests failing. The `scripts/retidy-prs`
helper fixes this by running `go mod tidy` across all modules on each affected PR
branch. Because the sweep relies on PR check status to decide what is mergeable,
this re-tidy **must** happen first — otherwise the sweep sees spurious failures and
either skips mergeable PRs or churns on them. This is a known brittle spot for this
project, so the pre-check below is mandatory.

## Project-specific

Run this **before** invoking `zed:dependabot-sweep`:

1. Find the open Dependabot PRs that currently have failing checks:

   ```bash
   gh pr list --author "app/dependabot" --state open \
     --json number,headRefName,statusCheckRollup \
     --jq '.[] | select([.statusCheckRollup[]? | select(.conclusion=="FAILURE")] | length > 0) | "\(.number) \(.headRefName)"'
   ```

2. If any are listed, run the project's re-tidy sweep to synchronize `go.mod` /
   `go.sum` across all modules on those branches:

   ```bash
   ./scripts/retidy-prs
   ```

   `retidy-prs` itself selects PRs with failing tests and runs `scripts/retidy-pr`
   on each (rebasing onto the base, running `go mod tidy` in every module
   directory, and force-pushing with lease). Let it run to completion.

3. **Verify the re-tidy actually succeeded** before proceeding. Confirm
   `retidy-prs` exited 0 and did not report any `[ERROR]` lines. Then re-check that
   the previously-failing Dependabot PRs either now have passing/pending checks or
   are failing for a reason unrelated to go.mod sync (re-run the `gh pr list`
   command from step 1). Do **not** continue to the sweep while a PR is still
   failing purely because its modules are out of tidy — investigate and re-run
   `./scripts/retidy-pr <branch>` for that branch first.

   If a re-tidy genuinely cannot be completed for some PR (e.g. an unresolvable
   rebase conflict), note which PR and why, and surface it in the final report
   rather than silently proceeding.

Only once the failing Dependabot PRs have been re-tidied and verified should you
continue to the standard sweep below.

## Steps

1. Complete the **Project-specific** `retidy-prs` verification above first.
2. Invoke the `zed:dependabot-sweep` skill and let it run to completion. It will
   unblock stuck Dependabot PRs, merge ready PRs, fix vulnerability alerts, update
   the changelog, and open a PR as appropriate for this repository.
3. Report what the re-tidy pre-check and the sweep did, including any PRs that
   could not be re-tidied.
