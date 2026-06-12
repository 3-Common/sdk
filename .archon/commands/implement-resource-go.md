---
description: Implement a new resource in the Go SDK (sdk-go) from the extracted OpenAPI slice, with examples and tests.
argument-hint: <resource/domain name, e.g. widgets>
---

# Implement the Go SDK resource

You are adding a new resource to the **Go** SDK only (`sdk-go/`). Follow the
`3common-sdk-conventions` skill for all layout, naming, codegen, and quality-gate
rules -- it is preloaded into your context.

## Canonical slug: use this, not the raw message
Read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as the
resource slug (always the lowercase OpenAPI path segment, the first path
segment after `/v1/`). Below,
**`<slug>`** means that value. Use `<slug>` verbatim for the conformance scenario
directory and the `call.resource` value. For Go **package, directory, and
identifier** names, a single-word slug is used as-is (package `contacts`, field
`Contacts`); a hyphenated slug (e.g. `payment-links`) uses a hyphen-free
lowercase package and matching directory (`paymentlinks`) plus PascalCase
exported field and type names (`PaymentLinks`), since Go package names cannot
contain hyphens.
Mirror the sibling's conventions. Do NOT derive paths or names from the raw invocation message, as it may be
capitalized or a whole sentence; only `domain` is canonical, and it is the exact
path the publish step stages.

## Inputs
- The API contract to wrap: `$ARTIFACTS_DIR/resource-spec.json` contains the `domain`
  slug plus every endpoint (method, path, params, request/response schemas),
  sliced from the canonical spec. This is your source of truth for what to wrap
  and how to call it.

## Scope
- Modify files **only under `sdk-go/`**. Do not touch `sdk-node/`,
  `sdk-python/`, `openapi/`, or any other SDK: those are handled separately and
  must not appear in this change.
- Do **not** run any `git` commands (no branch/add/commit/push). A later step
  owns all git and PR creation. Just leave your changes in the working tree.

## Steps
0. **Fetch dependencies**: `cd sdk-go && go mod download`.
1. **No codegen: hand-write the types.** The Go SDK does NOT use a generated
   layer: `sdk-go/generated/` is vestigial (no resource imports it), so you write
   all request/response types by hand in step 3. Do NOT run `make gen` or try to
   regenerate it (it may fail to run, which is expected and harmless, and CI never
   runs it).
2. **Study a sibling.** Read `sdk-go/resources/contacts/{client.go,types.go,client_test.go}` (or the closest-shaped sibling) and mirror its package structure, doc comments, and idioms, including both `New(cfg)` and `FromBackend(backend)` constructors.
3. **Write the resource** under `sdk-go/resources/<slug>/`: `types.go`, `client.go` (`package <slug>`) -- one method per endpoint in `resource-spec.json`.
4. **Mount it** in `sdk-go/client/api.go`: add the `<Name> *<slug>.Client` field to `type API struct` and assign `<Name>: <slug>.FromBackend(backend)` in `New` (`<Name>` = the capitalized slug).
5. **Add examples**: one directory per endpoint under `sdk-go/examples/<slug>/<op>/main.go` (snake_case dir, `package main`), each starting with `// Run with: go run ./examples/<slug>/<op>` and using the `"3co_your_api_key_here"` placeholder key. Mirror the sibling's example set.
6. **Write tests** in `sdk-go/resources/<slug>/client_test.go` (table-driven, like the sibling), covering every method and error path. Meet the gates in `.testcoverage.yml`: **>= 85% total, >= 80% per package, >= 75% per file**.
7. **Wire conformance.** The shared scenarios in `conformance/scenarios/<slug>/` are already generated and define the canonical method names/args your Go API must match. Create `sdk-go/conformance/dispatch_<slug>_test.go` (mirror `dispatch_contacts_test.go`) and register it: add a `case "<slug>":` to `dispatch()` in `sdk-go/conformance/runner_test.go`.
8. **Update the changelog**: add a bullet under `## [Unreleased]` -> `### Added` in `sdk-go/CHANGELOG.md` (create the `### Added` subheading if absent), mirroring the most recent resource entry's wording. See the conventions skill.
9. **Get to green** (run from `sdk-go/`): `gofumpt -w .`, then iterate until `make ci` is clean (vet + lint + race tests + coverage; the `conformance` package runs here too). Ensure examples compile (`go build ./examples/...`). Do not stop while anything is red.

## Output
Write a short summary to `$ARTIFACTS_DIR/go-summary.md`: the files you created,
the endpoints wrapped, and the final coverage number.
