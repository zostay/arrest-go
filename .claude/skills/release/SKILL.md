---
name: release
description: Cut a release of arrest-go — create a release/vX.Y.Z branch that the Prepare workflow dry-runs, merge it once checks pass, then tag vX.Y.Z to trigger the Release workflow, which also tags gin/vX.Y.Z. Invoke as `/release [version]`.
---

# Release

Cut a release of arrest-go. Two GitHub workflows validate the same things, so a
mistake is caught on the release branch (by `prepare.yaml`) before it can reach
the tag (`release.yaml`).

- `.github/workflows/prepare.yaml` — runs on push to `release/*`. Checks
  `version.txt`, `gin/go.mod`, the changelog heading, and that the embedded
  `arrest.Version` matches. This is the dry run.
- `.github/workflows/release.yaml` — runs on push of a `v*` tag. Repeats the
  same checks, pushes the `gin/vX.Y.Z` tag, creates the GitHub release from the
  changelog section, and un-drafts it.

`test.yaml` runs on every push, so the release branch is also tested normally.

## Two modules, two tags

The Gin integration is a separate Go module in `gin/`. Go resolves versions for
a module in a subdirectory from tags prefixed with that directory, so every
release is two tags: `vX.Y.Z` (root) and `gin/vX.Y.Z` (gin). You push only the
first; the Release workflow pushes the second at the same commit.

Because `gin/go.mod` has `replace github.com/zostay/arrest-go => ../`, its
`require` line for the root module is ignored locally and in CI but is what
consumers see. It must name the version being released, and both workflows
check that it does. The `gin/examples/polymorphic` module has the same shape
and is bumped for consistency, though nothing checks it.

## The four things the workflows enforce

1. **`version.txt` must contain the release version** (bare `X.Y.Z`).
   `arrest.Version` is embedded from it; the workflows run
   `go run ./internal/cmd/version` to confirm the two agree.
2. **`gin/go.mod` must `require github.com/zostay/arrest-go vX.Y.Z`.**
3. **The first line of `Changes.md` must be exactly `## X.Y.Z  YYYY-MM-DD`** —
   two spaces between version and date. A leftover `## Unreleased` section
   above it fails the check.
4. **That date must be today in `America/Chicago`** at the moment each
   workflow runs. If the Central date rolls over between merging the branch
   and pushing the tag, the Release workflow fails even though Prepare passed.
   See "If the date rolls over" below.

The release notes are everything between the first `## <digit>` heading and
the next one.

## Steps

### 1. Preflight

```bash
git status --porcelain          # must be clean
git rev-parse --abbrev-ref HEAD # must be master
git pull --ff-only
gh run list --branch master --limit 1   # master should be green
make test
```

Stop and tell the user if the tree is dirty or master is red. `Changes.md`
must have at least one `## Unreleased` section; if it does not, there is
nothing to cut.

### 2. Choose the version

```bash
cat version.txt
git tag --sort=-v:refname | head -3
```

If the user passed a version, use it. Otherwise propose the next one from the
unreleased entries — behaviour changes or new features mean a minor bump on
this 0.x project, pure fixes mean a patch bump — and **ask the user to confirm
before proceeding**. The version becomes a permanent public tag.

Bare `X.Y.Z` goes in `version.txt` and the changelog heading; `vX.Y.Z` is the
branch name, the tag, and the `gin/go.mod` require.

### 3. Create the release branch

```bash
git checkout -b release/vX.Y.Z
```

The branch name must contain the version; the workflows extract it with
`grep -Eo '[0-9]+\.[0-9]+\.[0-9]+.*$'`.

### 4. Update the version, modules, and changelog

This shell has `noclobber` set, so use `>|` when overwriting a file.

```bash
printf '%s\n' "X.Y.Z" >| version.txt
sed -i '' -E 's#^(\s*github.com/zostay/arrest-go) v[0-9].*$#\1 vX.Y.Z#' gin/go.mod
sed -i '' -E 's#^(\s*github.com/zostay/arrest-go) v[0-9].*$#\1 vX.Y.Z#; s#^(\s*github.com/zostay/arrest-go/gin) v[0-9].*$#\1 vX.Y.Z#' gin/examples/polymorphic/go.mod
make mod-tidy
```

Then consolidate every `## Unreleased` section in `Changes.md` into a single
`## X.Y.Z  <today>` section at the top, with the date from Central time:

```bash
TZ=America/Chicago date +%Y-%m-%d
```

Edit the bullets into release notes a consumer would want to read: drop noise,
group related entries, keep the wording user-facing. They are published
verbatim as the GitHub release body.

Verify by running the same checks the workflows run:

```bash
RELEASE_VERSION=X.Y.Z
grep -q "$RELEASE_VERSION" version.txt && echo "version.txt PASS"
grep -Eq "^\s*github.com/zostay/arrest-go v$RELEASE_VERSION\b" gin/go.mod && echo "gin/go.mod PASS"
date=$(TZ=America/Chicago date "+%Y-%m-%d")
[ "$(head -n1 Changes.md)" = "## $RELEASE_VERSION  $date" ] && echo "heading PASS"
grep -c '^## Unreleased' Changes.md    # must be 0
[ "$(go run ./internal/cmd/version)" = "$RELEASE_VERSION" ] && echo "embedded PASS"
make test
```

Preview the release body (the workflow's `sed` is GNU-only; use awk locally):

```bash
awk '/^## [0-9]/{n++; if(n==2) exit; next} n==1' Changes.md
```

### 5. Commit and push the release branch

```bash
git add version.txt Changes.md gin/go.mod gin/go.sum gin/examples/polymorphic/go.mod gin/examples/polymorphic/go.sum
git commit -m "chore(releng): version and changes"
git push -u origin release/vX.Y.Z
gh pr create --base master --title "Release vX.Y.Z" --body "<summary of the release>"
```

Use that exact commit message; it is the convention across projects.

### 6. Wait for the checks

```bash
gh pr checks <number> --watch
gh run list --branch release/vX.Y.Z --limit 5
```

**Prepare Release** is the dry run; if it fails, fix the cause on the branch
and push again. The three test jobs must pass too.

### 7. Merge

```bash
gh pr merge <number> --merge --delete-branch
git checkout master && git pull --ff-only
```

### 8. Tag the release

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

This triggers `release.yaml`. Do not push `gin/vX.Y.Z` yourself; the workflow
does it at the same commit.

### 9. Verify the release

```bash
gh run list --workflow=Release --limit 1
gh run watch <run-id>
gh release view vX.Y.Z
git fetch --tags && git tag --points-at vX.Y.Z    # must list gin/vX.Y.Z too
```

Confirm the release exists and is not a draft, and that both tags point at the
merge commit. Then confirm the module is actually consumable, since a green
workflow does not prove the proxy can see it:

```bash
cd "$(mktemp -d)" && go mod init probe >/dev/null
GOFLAGS=-mod=mod go get github.com/zostay/arrest-go@vX.Y.Z github.com/zostay/arrest-go/gin@vX.Y.Z
```

Check the release body matches the changelog section:

```bash
gh release view vX.Y.Z --json body --jq .body | sed '/^$/d' > /tmp/body.txt
awk '/^## [0-9]/{n++; if(n==2) exit; next} n==1' Changes.md | sed '/^$/d' | diff - /tmp/body.txt
```

### 10. Report

Report the version, the PR, both tags, and the release URL. Note anything that
needed a retry.

## If the date rolls over

If the Central date changes between the Prepare run and the tag push, the
Release workflow fails its changelog date check. Fix it on `master`:

1. Update the date in the `Changes.md` heading to the new Central date.
2. Commit to `master` via a PR, matching repo convention.
3. Delete and re-push the tag so it points at the corrected commit:

```bash
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
git tag vX.Y.Z
git push origin vX.Y.Z
```

If the failed run already pushed `gin/vX.Y.Z` or created a draft release,
delete those first (`git push origin :refs/tags/gin/vX.Y.Z`,
`gh release delete vX.Y.Z`) so the retry can recreate them cleanly.

## Notes

- Never push a `v*` tag without going through the release branch first. The
  tag is what publishes to the public; the branch is the only dry run.
- Do not amend or force-push a tag that has already produced a published
  release, and never move a tag the Go module proxy may have cached; cut a new
  patch version instead.
