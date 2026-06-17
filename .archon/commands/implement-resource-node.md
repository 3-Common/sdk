---
description: Implement a new resource in the Node SDK (sdk-node) from the extracted OpenAPI slice, with examples and tests.
argument-hint: <resource/domain name, e.g. widgets>
---

# Implement the Node SDK resource (the cross-language reference)

You are adding a new resource to the **Node** SDK only (`sdk-node/`). Follow the
`3common-sdk-conventions` skill for all layout, naming, codegen, and quality-gate
rules -- it is preloaded into your context.

**Node is the reference implementation.** It runs first, before any conformance
scenarios exist, because its types are generated from the canonical spec, so its
wire shapes are machine-checked against `openapi/` rather than hand-typed. The
method names and argument shapes you choose here BECOME the cross-language
contract: the next step derives the shared conformance scenarios from this
implementation, and the Python and Go SDKs are then written to match it. Choose
names and arg shapes deliberately, following the skill's conventions verbatim
(standard method names like `list`/`retrieve`/`create`/`update`/`listAutoPaginate`,
and the `call.args` id-naming rule). There are no scenarios to match yet -- you
are defining them.

## Canonical slug: use this, not the raw message
Read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as the
resource slug (always the lowercase OpenAPI path segment, the first path
segment after `/v1/`). Below,
**`<slug>`** means that value. Use `<slug>` verbatim for path-like names: the
resource directory, examples directory, conformance scenario directory, file
names, and the `call.resource` value. For **code identifiers**, derive the form
the sibling uses: a single-word slug is valid as-is, while a hyphenated slug
(e.g. `payment-links`) becomes camelCase for value identifiers (the `client`
property and factory: `paymentLinks`, `paymentLinksService`) and PascalCase for
type/interface names (`PaymentLinksService`). Do NOT derive
paths or names from the raw invocation message, as it may be capitalized or a whole
sentence; only `domain` is canonical, and it is the exact path the publish step
stages.

## Inputs
- The API contract to wrap: `$ARTIFACTS_DIR/resource-spec.json` contains the `domain`
  slug plus every endpoint (method, path, params, request/response schemas),
  sliced from the canonical spec. This is your source of truth for what to wrap
  and how to call it.

## Scope
- Modify files **only under `sdk-node/`**. Do not touch `sdk-python/`,
  `sdk-go/`, `openapi/`, or any other SDK: those are handled separately and
  must not appear in this change.
- Do **not** run any `git` commands (no branch/add/commit/push). A later step
  owns all git and PR creation. Just leave your changes in the working tree.

## Steps
0. **Install dependencies** -- this project uses **Yarn 1 Classic**, not npm (there is no `package-lock.json`; the worktree has no `node_modules`): `cd sdk-node && yarn install --frozen-lockfile`.
1. **Regenerate types from the spec** (idempotent): `cd sdk-node && yarn generate:types`. Confirm `<slug>`'s paths now resolve under `paths[...]` in `src/generated/types.ts`.
2. **Study a sibling.** Read `sdk-node/src/resources/contacts/{client.ts,types.ts,index.ts}` (or the sibling whose endpoint shape is closest to `<slug>`) and mirror its structure, TSDoc density, and idioms exactly.
3. **Write the resource** under `sdk-node/src/resources/<slug>/`: `types.ts`, `client.ts`, `index.ts` -- one service method per endpoint in `resource-spec.json`.
4. **Mount it**: add the property + assignment in `sdk-node/src/client.ts` (`ThreeCommon`), and `export * from './<slug>'` in `sdk-node/src/resources/index.ts`.
5. **Add examples**: one runnable file per endpoint under `sdk-node/examples/<slug>/<op>.ts` (kebab-case), mirroring the sibling's example set (include `error-handling` and an auto-paginate example if the domain paginates). Use the `"3co_your_api_key_here"` placeholder key.
6. **Write tests** under `sdk-node/test/` matching the sibling's test layout, covering every method and error path. Meet the **100%** coverage gate: `vitest.config.ts` enforces 100% lines/branches/functions/statements. Mirror how the sibling (e.g. `contacts`) reaches 100%; if a pure type-only re-export file genuinely can't be exercised, add it to the per-file exclude list in `vitest.config.ts`, as the `events` resource does.
7. **Wire conformance.** No scenarios exist for `<slug>` yet -- you are defining the contract they will encode. Still create `sdk-node/test/conformance/dispatch-<slug>.ts` (exporting `dispatch<Resource>`, mirroring `dispatch-contacts.ts`) and register it: import it and add a `case '<slug>':` to the `switch` in `sdk-node/test/conformance/runner.test.ts`. The dispatcher maps each standard method name + args (per the skill's conventions) to the client call. The conformance suite is inert for `<slug>` this run (no scenarios), but the next step derives scenarios that exercise exactly this dispatcher, so its method names and arg handling must follow the conventions precisely.
8. **Add a changeset** (Node uses changesets, do NOT hand-edit `CHANGELOG.md`). Create `sdk-node/.changeset/add-<slug>-resource.md` using the changeset format in the conventions skill's Changelogs section (frontmatter `"@3common/sdk": minor`, then a one-paragraph summary of the methods added). `.changeset/` holds only `README.md`/`config.json`: past changesets were consumed at release, so there's no prior changeset to copy; match the wording of the latest `sdk-node/CHANGELOG.md` entry for style.
9. **Get to green** (run from `sdk-node/`, via yarn, mirroring `.github/workflows/node.yml` exactly): `yarn format` to auto-format, then iterate until all pass -- `yarn lint`, `yarn format:check`, `yarn typecheck`, `yarn typecheck:build`, `yarn docs:check`, `yarn test:coverage`, `yarn build` (the conformance suite under `test/` runs as part of `test:coverage`). Do not stop while any is red.

## Output
Write a short summary to `$ARTIFACTS_DIR/node-summary.md`: the files you created,
the endpoints wrapped, and the final coverage number.
