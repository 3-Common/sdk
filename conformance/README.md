# Conformance scenarios

Cross-language behavioral tests that every SDK must pass identically. Each YAML file in [`scenarios/`](./scenarios) describes one SDK call: the inputs, the expected wire-level request(s), the mock response(s), and the expected return value or thrown error. Each language's test harness walks the directory recursively and asserts on every file.

If the Node SDK lowercases a header and the Go SDK doesn't, the YAML scenario fails on whichever SDK diverged. This is how we keep three SDKs aligned.

## Directory layout

Scenarios are grouped by resource so the set can grow without becoming a flat dump:

```
scenarios/
├── events/      # call.resource: events
│   ├── list-happy.yaml
│   ├── retrieve-happy.yaml
│   ├── error-401-unauthorized.yaml
│   └── …
└── invoices/    # call.resource: invoices
    ├── list-happy.yaml
    ├── create-happy.yaml
    ├── finalize-happy.yaml
    ├── error-409-finalize-conflict.yaml
    └── …
```

When you add a new resource (`subscriptions/`, `quotes/`, …), create a sibling subdirectory. The harnesses pick it up automatically — no runner changes needed unless the new resource introduces a method shape the dispatcher doesn't already handle.

## Scenario schema

### Common fields

```yaml
# Required: human-readable name.
name: list events — happy path

# Required: which SDK call to make. The harness translates the resource + method
# to the language-specific binding (e.g. `client.events.list(...)` in Node,
# `client.events.list(...)` in Python, `client.Events.List(ctx, ...)` in Go).
call:
  resource: events
  method: list | retrieve | update | listAutoPaginate
  args:                       # method-specific; the harness dispatches to the right shape
    status: open              # for list:    args is the params object
    pageSize: 10
    # id: evt_123             # for retrieve / update: args.id + args.params or args.body
    # body: { name: '...' }

# Optional: per-scenario client overrides.
client:
  apiVersion: '2026-04-29'    # default unless overridden
  telemetry: true             # default true
  maxRetries: 3               # default 3
```

### Single-call scenarios — happy path

```yaml
expectedRequest:
  method: GET
  path: /v1/events
  query:
    status: open
    pageSize: '10'            # query values are always stringified on the wire
  headers:                    # header keys are lowercased; assertion is "contains"
    authorization: 'Bearer 3co_test'
    threecommon-version: '2026-04-29'
  headersAbsent:              # optional: assert these are NOT present
    - threecommon-client-telemetry
  body: null                  # for GET; or { ... } for PATCH/POST

mockResponse:
  status: 200
  headers:
    x-request-id: req-list-001
  body:
    data: [...]
    hasMore: false

expectedResult:               # the SDK's return value (deep-equal subset match)
  data: [...]
  hasMore: false
```

### Single-call scenarios — error path

```yaml
expectedRequest: { ... }
mockResponse:
  status: 404
  body:
    error: { code: not_found, message: "..." }

expectedError:
  type: ThreeCommonNotFoundError    # SDK-specific class name; mapped per language
  code: not_found
  httpStatus: 404
  requestId: req-404-001
  retryAfterSeconds: 60             # only on rate-limit errors
  details: { ... }                  # optional
```

### Multi-call scenarios — retries and pagination

```yaml
exchanges:                    # ordered request/response pairs
  - request:
      method: GET
      path: /v1/events
      query: { page: '0' }
    response:
      status: 200
      body: { data: [...], hasMore: true }
  - request: { method: GET, path: /v1/events, query: { page: '1' } }
    response: { status: 200, body: { data: [...], hasMore: false } }

expectedAutoPaginated:        # for listAutoPaginate: expected sequence of items
  - { id: evt_1 }
  - { id: evt_2 }
  - { id: evt_3 }

# OR
expectedResult: { ... }       # for non-pagination retries

expectedCallCount: 2          # asserts the number of HTTP calls actually made
```

## Running

Each SDK runs the full scenario set as part of its merge-gate CI. The harness lives at:

- `sdk-node/test/conformance/runner.test.ts`
- `sdk-python/tests/test_conformance.py`
- `sdk-go/conformance/runner_test.go`

## Authoring guidance

- **Keep happy-path scenarios minimal.** Just enough fields in `mockResponse.body.data[*]` to satisfy the SDK's deserializer. Don't replicate every event field — only what the SDK actually needs.
- **Header keys are lowercase** in `expectedRequest.headers`. SDKs may emit canonical-case (`Authorization`) but the assertion compares lowercase.
- **Query values are strings** in `expectedRequest.query` because that's what `URLSearchParams` produces.
- **`expectedError.type`** uses Node's class name (`ThreeCommonNotFoundError`); Python and Go harnesses map to their language equivalents (`threecommon.error.NotFoundError`, `*threecommon.NotFoundError`).
- **`expectedResult`** uses subset matching — the SDK may return additional fields beyond what's asserted. Only the listed fields must equal exactly.
