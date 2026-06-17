---
description: Derive the shared cross-language conformance scenarios for a new SDK resource from the verified Node reference implementation.
argument-hint: <resource/domain name, e.g. widgets>
---

# Derive conformance scenarios from the Node reference

Create the shared, language-agnostic conformance scenario set for a new SDK
resource by capturing the behavior of the **already-implemented Node reference
SDK**. Node was implemented and passed its full gate first, and its types are
generated from the canonical spec, so it is the trustworthy source of the
contract. These scenarios become the equivalence gate that the Python and Go SDKs
are then written against, so they must faithfully encode what Node actually does
-- do NOT invent behavior. Follow the `3common-sdk-conventions` skill (Conformance
section), it is preloaded.

## Canonical slug: use this, not the raw message
First read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as
the resource slug (always the lowercase OpenAPI path segment, the first path
segment after `/v1/`).
Below, **`<slug>`** means that value. Use `<slug>` verbatim for the directory
name and as the `call.resource` value in every scenario. Do NOT derive paths or
names from the raw invocation message, as it may be capitalized or a whole sentence
(e.g. a full request sentence); only `domain` is canonical, and it is the exact
path the downstream publish step stages.

## Inputs (the Node reference is your source of truth)
- The reference implementation:
  `sdk-node/src/resources/<slug>/{client.ts,types.ts,index.ts}`. The exported
  service methods, their argument shapes, the wire request each builds (method,
  path, query, body), and the response type each returns are the contract. Read
  `$ARTIFACTS_DIR/node-summary.md` for the endpoint list and the method names the
  reference settled on.
- The Node conformance dispatcher:
  `sdk-node/test/conformance/dispatch-<slug>.ts` shows how each `call.method` +
  `call.args` maps to a client call -- match its method names and arg keys exactly
  in your scenarios.
- Endpoints + schemas: `$ARTIFACTS_DIR/resource-spec.json` for the wire shapes the
  reference wraps (the reference is authoritative where they differ).
- Schema + authoring rules: read `conformance/README.md`.
- Template: study a sibling set whose shape is closest, e.g.
  `conformance/scenarios/contacts/` and `conformance/scenarios/invoices/`.

## What to produce
Write YAML files under **`conformance/scenarios/<slug>/`** ONLY. Do not modify any
`sdk-*/` source: the Node implementation and its dispatcher are already done, and
the Python and Go implementations happen in later steps. Follow
`conformance/README.md`'s schema and copy a sibling's structure.

Mirror the coverage depth of the sibling set:
- One **happy-path** scenario per endpoint in `resource-spec.json`, with
  `call.method`/`call.args` matching the Node dispatcher and the
  `mockResponse`/`expectedResult` matching what the Node service actually returns.
- **Error paths for resource-specific failures only** -- typically 404
  not-found, 409 conflict, 422/400 validation, and any domain-specific 4xx the
  endpoints define, matching the wire shapes in the spec. Do NOT author scenarios
  for generic auth/transport errors (401, 403, 429, 5xx) even if the spec lists
  them: they are shared client behavior, not per-resource, and no sibling set
  includes them. Match the closest sibling set's error coverage.
- **Pagination / auto-paginate** scenarios (multi-call `exchanges`) for any list
  endpoint that paginates.
- Header/version/telemetry scenarios only if the sibling set includes them.

## Self-validate against the reference (required)
The scenarios must pass against the Node reference before you finish -- this is
how you prove they encode real behavior and catch transcription errors. From
`sdk-node/` run the conformance subset (the worktree already has `node_modules`
from the Node implementation step; if not, run `yarn install --frozen-lockfile`
first):

    cd sdk-node && yarn vitest run test/conformance

Iterate on the scenarios until every one for `<slug>` passes. If a scenario cannot
be made to pass without changing Node, the scenario is wrong (Node is the
reference): fix the YAML, not the SDK. Do NOT edit `sdk-node/` to chase a green
result.

## Output
Write a short summary to `$ARTIFACTS_DIR/conformance-summary.md`: the slug, the
scenario files created, the canonical method names per endpoint (as taken from the
Node reference), and confirmation that the Node conformance subset passes.
