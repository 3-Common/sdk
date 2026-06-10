---
name: 3common-sdk-conventions
description: |
  Conventions for adding/maintaining a resource in the 3Common SDK monorepo
  (sdk-node, sdk-python, sdk-go). Covers the per-resource file layout, the
  hand-curated-wrapper-over-generated-types pattern, codegen commands, the
  per-endpoint examples, the lint/type/test gates each PR must pass, commit/PR
  conventions, and the rules about what must NOT be hand-edited. Load this
  before implementing a resource.
---

# 3Common SDK conventions

This monorepo ships one SDK per language under `sdk-node/`, `sdk-python/`, and
`sdk-go/`. All three wrap the same HTTP API and are driven by a single canonical
OpenAPI document at `openapi/spec.yaml` (and `openapi/spec.json`), which is
fetched from the live server — **never hand-edited**.

## The core pattern: hand-curated wrappers over generated types

Each SDK has a **generated** layer (types produced mechanically from the spec)
and a **hand-curated** layer (the ergonomic, public resource API). You only ever
write the hand-curated layer. The generated layer is regenerated from the spec.

| Language | Generated (do NOT hand-edit) | Hand-curated (you write this) |
|----------|------------------------------|-------------------------------|
| Node | `sdk-node/src/generated/types.ts` | `sdk-node/src/resources/<name>/` |
| Python | `sdk-python/src/threecommon/_generated/models.py` | `sdk-python/src/threecommon/<name>/` |
| Go | `sdk-go/generated/` | `sdk-go/resources/<name>/` |

**Always regenerate the generated layer from the spec before writing the
wrapper** — the new domain's types may not be present yet. The commands are
idempotent (no-op if already current):

- **Node:** `cd sdk-node && yarn generate:types`
- **Python:** `cd sdk-python && datamodel-codegen --input ../openapi/spec.yaml --input-file-type openapi --output src/threecommon/_generated/models.py --output-model-type pydantic_v2.BaseModel --target-python-version 3.10 --use-standard-collections --use-union-operator --use-double-quotes --field-constraints --use-schema-description --capitalise-enum-members --reuse-model --openapi-scopes paths schemas parameters`
- **Go:** `cd sdk-go && make gen`

## Per-resource file layout

Mirror an existing sibling resource exactly (e.g. `contacts`). Pick the sibling
whose shape is closest to the new domain (a read-only domain → mirror a
read-heavy resource; a CRUD domain → mirror `contacts`).

### Node (`sdk-node/src/resources/<name>/`)
- `client.ts` — exported `XService` interface (TSDoc on every method, `@public`)
  plus an internal `xService(http: HttpClient): XService` factory (`@internal`).
  Use `http.request<T>({ method, path, query, body, options })`. Single-item
  responses come wrapped — unwrap `.data` via the local `DetailEnvelope<T>`
  helper. List endpoints return `{ data, hasMore, pageNumber, pageSize }`.
  Provide `listAutoPaginate` via `createAutoPaginator` when the domain paginates.
  Guard id args with a `requireId` helper.
- `types.ts` — friendly aliases over `paths[...]` from `@/generated/types`
  (e.g. `paths['/v1/<name>/{id}']['get']['responses'][200]['content']['application/json']['data']`),
  plus hand-written param/response interfaces.
- `index.ts` — barrel re-exporting the service + all public types.
- **Mount:** import + declare a `public readonly <name>: XService` property +
  assign `this.<name> = xService(this.httpClient)` in `sdk-node/src/client.ts`
  (`ThreeCommon` class), and add `export * from './<name>'` to
  `sdk-node/src/resources/index.ts`.

### Python (`sdk-python/src/threecommon/<name>/`)
- `service.py` — both a sync `XService` and an async `AsyncXService`.
- `types.py` — hand-curated public types over `_generated/models.py`.
- `__init__.py` — barrel exporting the services + types with an explicit
  `__all__` tuple (alphabetized).
- **Mount in `sdk-python/src/threecommon/client.py`** in BOTH `ThreeCommon`
  (sync) and `AsyncThreeCommon` (async): import the services, declare the
  attribute, and assign it in `__init__`.

### Go (`sdk-go/resources/<name>/`)
- `client.go` — `package <name>` with a `Client` struct and both a
  `New(cfg threecommon.Config) (*Client, error)` constructor and a
  `FromBackend(backend *core.Client) *Client` constructor.
- `types.go` — hand-curated response shapes (request types come from
  `sdk-go/generated`).
- `client_test.go` — table-driven tests (see coverage gate below).
- **Mount in `sdk-go/client/api.go`:** add a `<Name> *<name>.Client` field to
  `type API struct` and assign `<Name>: <name>.FromBackend(backend)` in `New`.

## Examples — one per endpoint, per language

Every SDK ships a runnable example for **each operation** of a resource under
`sdk-<lang>/examples/<resource>/`. When you add a resource, add a complete set
of examples covering every endpoint you wrapped (plus an `error-handling` and,
where the domain paginates, an auto-paginate example — match what the sibling
resource provides). Use the placeholder API key style `"3co_your_api_key_here"`.

| Language | Layout | Naming | Run with |
|----------|--------|--------|----------|
| Node | one **file** per op: `sdk-node/examples/<resource>/<op>.ts` | kebab-case (`auto-paginate.ts`, `list-activity.ts`) | `yarn tsx examples/<resource>/<op>.ts` |
| Python | one **file** per op: `sdk-python/examples/<resource>/<op>.py` | snake_case, with `_sync`/`_async` variants where the sibling has them (`list_sync.py`, `list_async.py`, `auto_paginate_async.py`) | `python examples/<resource>/<op>.py` |
| Go | one **directory** per op: `sdk-go/examples/<resource>/<op>/main.go` | snake_case dir, `package main` | `go run ./examples/<resource>/<op>` |

Mirror the sibling resource's example set exactly — same operation coverage,
same file/dir naming style, same header comment convention (Go examples start
with `// Run with: go run ./examples/<resource>/<op>`).

## Conformance (REQUIRED for all three languages)

Cross-language behavioral tests live in `conformance/scenarios/<resource>/*.yaml`
— one shared, language-agnostic set per resource. Each scenario describes one
SDK call: inputs, the expected wire request(s), the mock response(s), and the
expected return value or thrown error. Every SDK runs the full set and must
behave identically. (Note: `conformance/README.md` is stale — it lists Python
and Go as "Planned". They are **live and required**; all three harnesses exist.)

Adding a resource requires BOTH of these:

1. **Shared scenarios** (do this once, not per language): create
   `conformance/scenarios/<resource>/` mirroring a sibling's coverage — a
   happy-path scenario per endpoint, the relevant error paths (404/409/422/etc.),
   and pagination/auto-paginate scenarios where the domain paginates. Follow the
   schema in `conformance/README.md` and copy the structure of a sibling set
   (e.g. `conformance/scenarios/contacts/`). The method names and arg shapes you
   choose here are the contract every language must match.

2. **A per-language dispatcher**, registered in that language's runner:

   | Language | Create dispatcher | Register it in |
   |----------|-------------------|----------------|
   | Node | `sdk-node/test/conformance/dispatch-<resource>.ts` | add a `case '<resource>':` to the `switch` in `runner.test.ts` (and import it) |
   | Python | `sdk-python/tests/_conformance/dispatch_<resource>.py` (with `dispatch_sync` and `dispatch_async`) | add `from _conformance import dispatch_<resource>` and an `if resource == "<resource>"` branch in BOTH `_dispatch_sync` and `_dispatch_async` in `tests/test_conformance.py` |
   | Go | `sdk-go/conformance/dispatch_<resource>_test.go` | add a `case "<resource>":` to `dispatch()` in `runner_test.go` |

The conformance suite is part of each language's normal test run (Node: under
`test/` → `yarn test`; Python: `tests/test_conformance.py` → `pytest`; Go: the
`sdk-go/conformance` package → `go test ./...` / `make ci`), so once the
scenarios and dispatcher exist, the gates below exercise them automatically.

## Quality gates (every PR must pass these — match CI)

Coverage gates are enforced and a drop fails the PR — and they **differ per
language**: **Node = 100%** (lines/branches/functions/statements, via
`sdk-node/vitest.config.ts`, minus its per-file exclude list); **Python = 90%**
total (`pytest --cov-fail-under=90`); **Go = 85% total** plus **80% per package**
and **75% per file** (`sdk-go/.testcoverage.yml`). Write tests alongside the
resource. (Note: the top-level `CONTRIBUTING.md` says "90% Node" — it is stale;
the config files above are authoritative.)

| Language | Lint | Type-check | Test (+coverage) |
|----------|------|-----------|------------------|
| Node | `yarn lint` + `yarn format:check` | `yarn typecheck` + `yarn typecheck:build` | `yarn test:coverage` |
| Python | `ruff check . && ruff format --check .` | `mypy src/threecommon tests && pyright src/threecommon` | `pytest --cov=src/threecommon --cov-fail-under=90` |
| Go | `golangci-lint run` (strict preset) + `gofumpt -l .` | compilation is the type-check (`go vet` / `go test`) | `make coverage` |

Python strictness is configured in `pyproject.toml` (mypy `strict = true`,
pyright `typeCheckingMode = standard`, a curated ruff `select`), so run the bare
commands above — do NOT add `--strict` (pyright has no such flag; mypy already
gets it from config) or `--select=ALL` (it overrides the project's rule policy
and fails existing code). Scope mypy to `src/threecommon tests`, never `.`,
which pulls in `examples/` and its reused module names. These mirror
`.github/workflows/python.yml`.

Node uses **Yarn 1 Classic** — install with `yarn install --frozen-lockfile` and
run scripts as `yarn <script>`, never npm (there is no `package-lock.json`). Its
full gate set is `yarn lint`, `yarn format:check`, `yarn typecheck`,
`yarn typecheck:build`, `yarn docs:check`, `yarn test:coverage`, `yarn build`,
mirroring `.github/workflows/node.yml`.

For Go, `make ci` runs vet + lint + race tests + coverage in one shot.
Run formatters before committing: Node `yarn format`, Python `ruff format .`,
Go `gofumpt -w .`.

**Examples are part of the gate too.** For Go they are `package main` programs
under the module, so `make ci` (which runs on `./...`) already vets, lints, and
compiles them; the Go gate additionally runs `go build ./examples/...` as an
explicit guard, since `PKGS` is overridable. For Node and Python, examples are
covered by the same lint + type-check commands above — make sure they pass.

## Changelogs (mechanism differs per language)

Each SDK keeps its own changelog under `sdk-<lang>/CHANGELOG.md`. The top-level
`CHANGELOG.md` is a static index — do not edit it. Adding a resource is an
**Added** / minor change. Record it for every language you touch:

- **Node — changesets (do NOT hand-edit `sdk-node/CHANGELOG.md`).** Add a new
  file `sdk-node/.changeset/add-<resource>-resource.md`. The CHANGELOG is
  generated from changesets at release time. Format:
  ```markdown
  ---
  "@3common/sdk": minor
  ---

  Add the `<resource>` resource. <one-paragraph summary of the methods added>.
  ```
- **Python — `sdk-python/CHANGELOG.md`** (Keep a Changelog). Add a bullet under
  the existing `## [Unreleased]` → `### Added` (create the `### Added`
  subheading if missing). Mention both sync and async surfaces.
- **Go — `sdk-go/CHANGELOG.md`** (Keep a Changelog). Add a bullet under
  `## [Unreleased]` → `### Added`.

Match the wording/structure of the most recent resource entry (e.g. the
`entitlements` entry) so the changelog reads consistently.

## Commit & PR conventions

- **Conventional Commits**, one scope per language: `feat(node):`,
  `feat(python):`, `feat(go):`. Example: `feat(node): add widgets resource`.
- A new resource ships as **one atomic PR** containing all three languages —
  per-language commits (`feat(node):`, `feat(python):`, `feat(go):`) plus a
  `test(conformance):` commit for the shared scenarios. They must land together:
  the cross-language conformance scenarios can't pass CI until all three
  language dispatchers coexist. (The SDKs are still *released* independently —
  releases are driven by per-language changesets/tags, not PR boundaries.)
- Open PRs as **drafts**; a maintainer reviews and marks ready.
- Secret scanning (`gitleaks`) runs in CI — never commit secrets or `.env`.

## Hard rules

- Do **not** edit `openapi/spec.{yaml,json}` or any generated layer by hand.
- Do **not** touch SDKs other than the one you are implementing — keep each
  language's change confined to its own `sdk-<lang>/` directory so it forms a
  clean, isolated per-language commit within the resource's single PR.
- Reference an existing sibling resource for naming, formatting, and structure
  so the new resource is indistinguishable in style from the rest of the SDK.
