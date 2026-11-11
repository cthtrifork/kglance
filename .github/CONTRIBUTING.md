# Contributing

1. Branch from `main`: `git checkout -b <topic> origin/main`.
2. Make the change with tests.
3. Run what CI runs: `make check` (gofmt, `go mod tidy -diff`, vet, tests), plus `make lint` for the linters.
4. Update docs and the README for user-visible changes.
5. Open a PR using the template. The title is `type(scope): summary` ([Conventional Commits](https://www.conventionalcommits.org/)); see [`docs/releasing.md`](../docs/releasing.md).

Standing rules and the layout: [`AGENTS.md`](../AGENTS.md).
