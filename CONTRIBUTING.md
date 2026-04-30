# Contributing

Thanks for your interest in contributing to the 3Common SDK. This document covers the development workflow.

## Quick start

```bash
git clone https://github.com/3-Common/sdk.git
cd sdk
pre-commit install
```

The repository contains one SDK per language under `sdk-*/`. Each has its own README with build and test instructions.

## Workflow

1. Open an issue describing the bug or feature before sending a PR for non-trivial changes.
2. Fork, branch off `main`, push, and open a pull request.
3. Each PR must pass:
   - All language-specific lint, type-check, and test jobs in CI.
   - Secret scanning (`gitleaks`).
   - Code review by a maintainer.
4. Coverage gates: ≥ 90% line + branch for Node and Python; ≥ 85% for Go. Drops fail the PR.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/). Examples:

- `feat(node): add events.listAutoPaginate`
- `fix(python): handle 429 with no Retry-After header`
- `docs: update install snippet`
- `chore(deps): bump httpx to 0.27`

Scopes are typically `node`, `python`, `go`, `openapi`, `ci`, `docs`.

## Tests

- **Unit + mocked-integration + contract**: run on every PR for every language.
- **Live smoke**: maintainer-only, runs against the production API before release tags.

## Code style

- TypeScript: `eslint` + Prettier; strict `tsconfig`.
- Python: `ruff format` and `ruff check --select=ALL`; `mypy --strict` and `pyright --strict`.
- Go: `gofumpt`, `golangci-lint` with the strict preset.

## Reporting security issues

Do **not** open a public issue. See [SECURITY.md](./SECURITY.md).

## License

By contributing, you agree your contributions are licensed under the [MIT License](./LICENSE).
