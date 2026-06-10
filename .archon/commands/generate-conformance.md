---
description: Generate the shared cross-language conformance scenarios for a new SDK resource from the extracted OpenAPI slice.
argument-hint: <resource/domain name, e.g. widgets>
---

# Generate conformance scenarios

Create the shared, language-agnostic conformance scenario set for a new SDK
resource. These scenarios are the behavioral contract that the Node, Python, and
Go SDKs will each implement and be tested against, so they must be decided
**before** implementation. Follow the `3common-sdk-conventions` skill
(Conformance section) — it is preloaded.

## Canonical slug — use this, not the raw message
First read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as
the resource slug (always the lowercase OpenAPI path segment — the first path
segment after `/v1/`).
Below, **`<slug>`** means that value. Use `<slug>` verbatim for the directory
name and as the `call.resource` value in every scenario. Do NOT derive paths or
names from the raw invocation message — it may be capitalized or a whole sentence
(e.g. a full request sentence); only `domain` is canonical, and it is the exact
path the downstream publish step stages.

## Inputs
- Endpoints to cover + the slug: `$ARTIFACTS_DIR/resource-spec.json` (the
  `domain` field plus every endpoint, with methods, params, request/response
  schemas).
- Schema + authoring rules: read `conformance/README.md`.
- Template: study a sibling set whose shape is closest, e.g.
  `conformance/scenarios/contacts/` and `conformance/scenarios/invoices/`.

## What to produce
Write YAML files under **`conformance/scenarios/<slug>/`** ONLY. Do not touch any
`sdk-*/` directory or the OpenAPI spec — implementation and per-language
dispatchers happen in later steps.

Mirror the coverage depth of the sibling set:
- One **happy-path** scenario per endpoint in `resource-spec.json`.
- The relevant **error paths** the endpoints can produce (e.g. 404 not-found,
  409 conflict, 422 validation), matching the wire shapes in the spec.
- **Pagination / auto-paginate** scenarios (multi-call `exchanges`) for any list
  endpoint that paginates.
- Header/version/telemetry scenarios only if the sibling set includes them.

## Critical: you are defining the API surface
The `call.resource` (= `<slug>`), `call.method`, and `call.args` you write are
the canonical cross-language API for this resource. Choose method names and arg
shapes consistent with sibling resources (`list`, `retrieve`, `create`,
`update`, `delete`, `listAutoPaginate`, …). The three implementations will be
written to match exactly what you specify here. Be deliberate and consistent.

## Output
Write a short summary to `$ARTIFACTS_DIR/conformance-summary.md`: the slug, the
scenario files created, and the canonical method names chosen per endpoint.
