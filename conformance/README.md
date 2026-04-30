# Conformance scenarios

Cross-language behavioral tests that every SDK must pass identically. Each scenario is a YAML file describing one request the SDK should make and how it should handle the response. Each language's test suite loads every file in [`scenarios/`](./scenarios) and asserts.

If the Node SDK lowercases a header and the Go SDK doesn't, the YAML scenario fails on whichever SDK diverged. This is how we keep three SDKs aligned.

## Scenario format

```yaml
# Identifier used in test output.
name: list events with status filter

# The SDK call to make. The test harness translates this to the language-specific
# method name (e.g. events.list / events.list / Events.List).
call:
  resource: events
  method: list
  args:
    status: open
    page_size: 10

# What the harness should observe on the wire after the SDK call.
expectedRequest:
  method: GET
  path: /v1/events
  query:
    status: open
    pageSize: "10"
  headers:
    Authorization: "Bearer 3co_test"
    Threecommon-Version: "2026-04-29"

# The mock response the harness should return to the SDK.
mockResponse:
  status: 200
  headers:
    X-Request-ID: req-test-001
    Content-Type: application/json
  body:
    data: []
    hasMore: false

# What the SDK should return to the caller.
expectedResult:
  data: []
  hasMore: false
```

Error scenarios use the same shape; `expectedResult` is replaced by `expectedError`:

```yaml
name: 404 on missing event
call:
  resource: events
  method: retrieve
  args: { id: "evt_missing" }
expectedRequest:
  method: GET
  path: /v1/events/evt_missing
mockResponse:
  status: 404
  body:
    error:
      code: not_found
      message: "Event not found"
expectedError:
  type: NotFoundError
  code: not_found
  httpStatus: 404
```

## Running

Each SDK runs the full scenario set as part of its merge-gate CI. See:

- `sdk-node/test/conformance/runner.test.ts`
- `sdk-python/tests/conformance/test_runner.py`
- `sdk-go/conformance_test.go`
