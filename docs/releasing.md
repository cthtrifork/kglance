# Releasing

A release is two files on a GitHub release: the `kglance` binary and a Fedora `.rpm`, plus `SHA256SUMS`, a CycloneDX SBOM and build-provenance attestations.

## Cut a release

**Actions → Tag and release → Run workflow** (on `main`), or:

```sh
gh workflow run tag-release.yml                    # next version, from the merged PR titles
gh workflow run tag-release.yml -f tag=v1.0.0      # an exact tag
```

The workflow refuses to run anywhere but `main`, works out the version (the `tag` input, otherwise Release Drafter's), stops if the tag exists, and calls `release.yml` with that version and the commit it started on. `release.yml` re-runs vet, tests and `govulncheck`, builds on Fedora 44, packages the rpm, attests everything and publishes the release. No tag is pushed from outside: GitHub creates it when the release is published, so a failed build leaves nothing to clean up.

A dry run (`release.yml` dispatched on a branch) builds everything but publishes nothing.

## Release notes

[Release Drafter](https://github.com/release-drafter/release-drafter) writes the notes and the next version from PR titles, using [`.github/release-drafter.yml`](../.github/release-drafter.yml). A PR title is all it needs: `type(scope)!: summary` ([Conventional Commits](https://www.conventionalcommits.org/)). Write the effect a user sees, not the diff.

| Type | Section | Next version |
|---|---|---|
| any type with `!` | Breaking changes | major |
| `feat` | New features | minor |
| `security`, `fix`, `perf`, `docs`, `deps`, `ci`, `build`, `refactor`, `test`, `chore`, `revert` | their own sections | patch |

`pr-title.yml` fails a PR whose title does not match and labels it from the type. When you change a type, change `pr-title.yml`, the categories in `.github/release-drafter.yml` and this table together. Make **check title** a required status check in the branch rules (a workflow cannot do that).

## Repository settings (once, by the owner)

- Branch rules on `main`: require a pull request, and the checks `lint`, `test`, `build (Fedora 44)` and `check title`; squash merge only.
- Create the Renovate GitHub App and add the `RENOVATE_APP_CLIENT_ID` variable and `RENOVATE_APP_PRIVATE_KEY` secret (see [ci](ci.md)); also require `no-screenshots`.
- Allow GitHub Actions to create releases and attestations (the defaults are enough).
