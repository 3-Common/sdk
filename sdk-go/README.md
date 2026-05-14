# `github.com/3-Common/sdk/sdk-go`

[![pkg.go.dev](https://pkg.go.dev/badge/github.com/3-Common/sdk/sdk-go.svg)](https://pkg.go.dev/github.com/3-Common/sdk/sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/3-Common/sdk/sdk-go)](https://goreportcard.com/report/github.com/3-Common/sdk/sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/go-%3E%3D1.22-brightgreen)](https://go.dev)

Official Go client for the 3Common Public API.

## Install

```bash
go get github.com/3-Common/sdk/sdk-go
```

Requires **Go ≥ 1.22**.

## Quick start

```go
package main

import (
    "context"
    "log"

    threecommon "github.com/3-Common/sdk/sdk-go"
    "github.com/3-Common/sdk/sdk-go/client"
    "github.com/3-Common/sdk/sdk-go/event"
)

func main() {
    api, err := client.New(threecommon.Config{APIKey: "3co_..."})
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List
    pageSize := 50
    result, err := api.Events.List(ctx, &event.ListParams{
        Status:   event.StatusOpen,
        PageSize: &pageSize,
    })
    if err != nil {
        log.Fatal(err)
    }
    _ = result.Data

    // Retrieve
    ev, err := api.Events.Retrieve(ctx, "evt_123", nil)
    if err != nil {
        log.Fatal(err)
    }
    _ = ev

    // Update
    updated, err := api.Events.Update(ctx, "evt_123", &event.UpdateParams{
        Name: threecommon.String("New name"),
    })
    if err != nil {
        log.Fatal(err)
    }
    _ = updated

    // Auto-paginate
    iter := api.Events.ListAutoPaginate(ctx, &event.ListParams{Status: event.StatusOpen})
    for iter.Next() {
        _ = iter.Current()
    }
    if err := iter.Err(); err != nil {
        log.Fatal(err)
    }
}
```

Generate API keys in the 3Common organizer dashboard (`Settings -> API Keys`). The key may also be supplied via the `THREECOMMON_API_KEY` environment variable.

## Configuration

```go
api, err := client.New(threecommon.Config{
    APIKey:     "3co_...",                  // required (or via env var)
    BaseURL:    "https://api.3common.com",  // default
    APIVersion: "2026-04-29",               // pinned API version
    Timeout:    30 * time.Second,           // per-request deadline
    MaxRetries: threecommon.Int(3),         // automatic retries on 408/425/429/5xx
    RetryDelay: threecommon.RetryDelay{
        Initial: 500 * time.Millisecond,
        Max:     8 * time.Second,
        Jitter:  true,
    },
    HTTPClient: &http.Client{Timeout: 60 * time.Second},  // override transport
    Logger:     myLogger,                                 // optional debug logger
    Telemetry:  threecommon.Bool(true),                   // opt-out of anonymous telemetry
})
```

## Error handling

Every error returned by the SDK wraps a `*threecommon.APIError` and is one of the typed subtypes. Branch with `errors.As`:

```go
import "errors"

_, err := api.Events.Retrieve(ctx, "evt_missing", nil)
if err != nil {
    var notFound *threecommon.NotFoundError
    var auth *threecommon.AuthError
    var rate *threecommon.RateLimitError

    switch {
    case errors.As(err, &notFound):
        // 404
    case errors.As(err, &auth):
        // 401 — bad or expired API key
    case errors.As(err, &rate):
        // 429 — rate.RetryAfter tells you when to retry
    default:
        return err
    }
}
```

Every error carries `Code`, `Message`, `HTTPStatus`, `RequestID`, `Details`, and `RawResponse`. The default `Error()` format includes the request ID for log correlation:

```
[not_found] Event evt_missing not found (request_id=req-dfx-abc)
```

## Pagination

Two flavors:

```go
// One page at a time
result, err := api.Events.List(ctx, &event.ListParams{PageSize: threecommon.Int(50)})

// All pages, lazy
iter := api.Events.ListAutoPaginate(ctx, nil)
for iter.Next() {
    ev := iter.Current()
    _ = ev
}
if err := iter.Err(); err != nil {
    return err
}
```

Go 1.23+ the iterator also support range-over-func variant:

```go
for ev, err := range api.Events.ListAutoPaginate(ctx, nil).All() {
    if err != nil {
        return err
    }
    _ = ev
}
```

## Filters

The `filters` subpackage provides a typed builder for the API's `filters` query parameter:

```go
import "github.com/3-Common/sdk/sdk-go/filters"

f := filters.And(
    filters.Field("status").IsAnyOf("open"),
    filters.Field("ticketSum").IsGreaterThan(10),
)

params := (&event.ListParams{}).FilterWith(f)
api.Events.List(ctx, params)
```

The full operator set is enumerated in `filters/types.go`.

## Retries

Idempotent methods (`GET`, `PATCH`, `PUT`) retry automatically on `408`, `425`, `429`, `500`, `502`, `503`, `504` and on network errors. Backoff is exponential with full jitter, capped at `RetryDelay.Max`. The SDK honors a server-provided `Retry-After` header on `429`.

`POST` and `DELETE` do not retry by default; pass an `Idempotency-Key` via per-request options to opt in (forward-compat — no v1 endpoints currently use this).

## Telemetry

The SDK sends a small, anonymized `Threecommon-Client-Telemetry` header on every request (SDK version, language, last-request latency). This helps debug performance reports from real customers without instrumenting their code. Disable globally:

```go
api, _ := client.New(threecommon.Config{APIKey: "...", Telemetry: threecommon.Bool(false)})
```

Or at runtime:

```go
api.DisableTelemetry()
```

The header never contains your API key, request bodies, or response bodies.

## Repository layout

The package layout mirrors the Node SDK so behavior stays in lockstep:

```
sdk-go/
├── doc.go              # package overview for pkg.go.dev
├── config.go           # Config + Logger + RetryDelay
├── errors_base.go      # *APIError + Error() + Unwrap()
├── errors_types.go     # typed subtypes: AuthError, NotFoundError, ...
├── helpers.go          # String / Int / Int64 / Bool / Float64 pointer helpers
├── version.go          # SDK version constant
├── api_version.go      # pinned API version + path
├── client/             # aggregator: client.New(cfg) → *client.API
├── resources/events/   # events resource (client + types)
├── filters/            # typed filter builder
├── pagination/         # generic Iter[T] auto-paginator
├── internal/core/      # HTTP transport (not user-importable)
└── generated/          # oapi-codegen output (re-run via `make gen`)
```

## Versioning

The SDK follows SemVer. The pinned **API version** (sent as `Threecommon-Version`) is independent — the API can evolve without breaking already-deployed SDKs. Bump `APIVersion` to opt into newer server behavior.

Module path: `github.com/3-Common/sdk/sdk-go`. Major bumps follow the standard Go major-version suffix convention (`/v2` from v2.0.0 onward).

## Examples

End-to-end runnable examples live under [`examples/events/`](./examples/events/):

```bash
go run ./examples/events/list
go run ./examples/events/retrieve
go run ./examples/events/update
go run ./examples/events/auto_paginate
go run ./examples/events/error_handling
go run ./examples/events/filters
```

Replace `3co_your_api_key_here` and `evt_replace_with_real_id` with real values before running.

## Contributing

See the [repository CONTRIBUTING guide](https://github.com/3-Common/sdk/blob/main/CONTRIBUTING.md). Issues and PRs welcome.

## License

[MIT](./LICENSE)
