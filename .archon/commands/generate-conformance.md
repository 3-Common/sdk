---
description: Generate the shared cross-language conformance scenarios for a new SDK resource from the extracted OpenAPI slice.
argument-hint: <resource/domain name, e.g. widgets>
---

# Generate conformance scenarios

Create the shared, language-agnostic conformance scenario set for a new SDK
resource. These scenarios are the behavioral contract that the Node, Python, and
Go SDKs will each implement and be tested against, so they must be decided
**before** implementation. Follow the `3common-sdk-conventions` skill
(Conformance section), it is preloaded.

## Canonical slug: use this, not the raw message
First read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as
the resource slug (always the lowercase OpenAPI path segment, the first path
segment after `/v1/`).
Below, **`<slug>`** means that value. Use `<slug>` verbatim for the directory
name and as the `call.resource` value in every scenario. Do NOT derive paths or
names from the raw invocation message, as it may be capitalized or a whole sentence
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
`sdk-*/` directory or the OpenAPI spec -- implementation and per-language
dispatchers happen in later steps.

Mirror the coverage depth of the sibling set:
- One **happy-path** scenario per endpoint in `resource-spec.json`.
- **Error paths for resource-specific failures only** -- typically 404
  not-found, 409 conflict, 422/400 validation, and any domain-specific 4xx the
  endpoints define, matching the wire shapes in the spec. Do NOT author scenarios
  for generic auth/transport errors (401, 403, 429, 5xx) even if the spec lists
  them: they are shared client behavior, not per-resource, and no sibling set
  includes them. Match the closest sibling set's error coverage.
- **Pagination / auto-paginate** scenarios (multi-call `exchanges`) for any list
  endpoint that paginates.
- Header/version/telemetry scenarios only if the sibling set includes them.

## Critical: you are defining the API surface
The `call.resource` (= `<slug>`), `call.method`, and `call.args` you write are
the canonical cross-language API for this resource. Choose method names and arg
shapes consistent with sibling resources (`list`, `retrieve`, `create`,
`update`, `delete`, `listAutoPaginate`, ...). Name `call.args` ids per the
skill's rule: the resource's own id is `id`, and every deeper path id is named
after its segment (`elementId`, `targetElementId`, ...) -- never a `<resource>Id`
form for the resource's own id. The three implementations will be written to
match exactly what you specify here. Be deliberate and consistent.

## Output
Write a short summary to `$ARTIFACTS_DIR/conformance-summary.md`: the slug, the
scenario files created, and the canonical method names chosen per endpoint.
