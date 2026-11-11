# CI

| Workflow | Runs on | What |
|---|---|---|
| `ci.yml` | every PR, push to `main` | `lint` (gofmt, tidy, bindings up to date, yamllint, golangci-lint, Biome, knip), `test` (vet, `go test -race`, frontend tests), `build (Fedora 44)` (binary, rpm, rpm installs, desktop files valid), `no-screenshots` |
| `screenshots.yml` | PRs and pushes to `main` that touch `frontend/**` | Renders the popup in demo mode; PR comment on PRs, a draft preview PR from `main` |
| `security.yml` | Go or dependency changes, weekly | govulncheck, Trivy filesystem scan, `npm audit` |
| `actions-lint.yml` | `.github/**` changes, weekly | zizmor, pinned actions, Renovate config validation |
| `pr-title.yml` | PRs | Conventional Commits title check and label |
| `release-drafter.yml` | push to `main` | Keeps the draft release notes and next version |
| `renovate.yml` | schedule, and PRs that change its config | Self-hosted Renovate; a dry run on config PRs |
| `update-branches.yml` | push to `main` | Merges `main` into the owner's open PRs that are behind |
| `tag-release.yml`, `release.yml` | manual | Cut a release, see [releasing](releasing.md) |

Actions are pinned by commit SHA with a version comment, and the Fedora job container is pinned by digest.

## Repository variable and secret

Everything that needs more than the default `GITHUB_TOKEN` uses one GitHub App, the one Renovate runs as. Add these in Settings, Secrets and variables, Actions:

| Kind | Name | Value |
|---|---|---|
| Variable | `RENOVATE_APP_CLIENT_ID` | The App's client ID (the `Iv...` string on the App's settings page) |
| Secret | `RENOVATE_APP_PRIVATE_KEY` | The App's private key (the whole downloaded `.pem`) |

Used by `renovate.yml`, `update-branches.yml` and the push-to-`main` part of `screenshots.yml`. Without them, Renovate and update-branches fail on their token step, and `screenshots.yml` still renders but does not push the preview. Nothing else needs a secret.

Create the App (Settings, Developer settings, GitHub Apps, New GitHub App), no webhook, install it on this repository only. Repository permissions: Contents, Pull requests, Issues, Workflows and Commit statuses read and write; Checks and Dependabot alerts read. Also turn on Settings, Actions, General, "Allow GitHub Actions to create and approve pull requests" (the screenshots preview PR is opened by the App, but Renovate's dry run and others read PRs through it).

Why an App and not `GITHUB_TOKEN`: PRs and pushes made with `GITHUB_TOKEN` start no other workflows, so CI would never report on Renovate's PRs (automerge would have no checks to wait for) or on the screenshots preview branch.

## Renovate

Self-hosted ([renovatebot/github-action](https://github.com/renovatebot/github-action)), config in `.github/renovate.json5`.

- Updates Go modules, the frontend npm packages, SHA-pinned Actions, the Fedora container digest, the Go and Node versions in `.tool-versions`, and anything marked `# renovate: datasource=<ds> depName=<name>` (the Makefile tool versions, yamllint, zizmor, Trivy, govulncheck, the Renovate version).
- A release must be 2 days old first. Patch and digest updates automerge once CI is green (so require the CI checks in the branch rules); minors are grouped per ecosystem; Go, Node, Wails and the Fedora release are reviewed by hand. Vite 8 is held back (its native bundler crashes on CPUs without AVX2) and Fedora stays on 44 until you decide otherwise.
- Runs every two hours on `main` and by hand. A PR that changes `renovate.json5` or `renovate.yml` gets a dry run on its own branch: nothing is written, and the job summary lists what Renovate would do.
- Commits go through the API as the App (`platformCommit`), so they are verified.

## Screenshots

`screenshots.yml` renders the popup in every demo state (battery, charging, low, no battery) in light and one of them in dark, with Playwright against the Vite dev server. The page only accepts `?demo=...` in the dev server, and the clock is fixed, so the images are stable. `make screenshots` does the same locally.

- **Pull request:** each image is uploaded as its own unzipped artifact and all are embedded in one sticky PR comment that is updated on every push. Viewers must be logged in, and the images expire after 14 days. Fork PRs get the artifacts but no comment.
- **Push to `main`:** the images are force-pushed as one commit on top of `main` to `bot/update-screenshots`, and one draft PR "Screenshots preview (do not merge)" is kept open from it, so the current pictures can be browsed in its Files changed tab. It is never merged; `no-screenshots` in `ci.yml` fails on it (add that check to the required checks).
- **Adding or removing a state** in `frontend/scripts/screenshots.mjs` needs a matching `Upload <name>` step (id `shot-<name>`) in `screenshots.yml`, because an action cannot be looped; the comment step fails with a message when the counts differ.

## update-branches

On every push to `main`, each open, non-draft, same-repository PR by the owner that is behind `main` gets `main` merged into it through the `update-branch` API, so CI runs again and the PR stays mergeable under "Require branches to be up to date". PRs by others, `renovate/*`, `bot/*`, forks and drafts are skipped.
