# `@3common/sdk`

[![npm](https://img.shields.io/npm/v/@3common/sdk.svg)](https://www.npmjs.com/package/@3common/sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Node](https://img.shields.io/badge/node-%3E%3D20-brightgreen)](https://nodejs.org)

Official Node.js / TypeScript client for the 3Common Public API.

## Install

```bash
npm install @3common/sdk
# or
pnpm add @3common/sdk
# or
yarn add @3common/sdk
```

Requires **Node.js ≥ 20**. Targets ESM and CJS via dual-emit; TypeScript types are bundled.

## Quick start

```ts
import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: process.env.THREECOMMON_API_KEY,
})

const { data, hasMore } = await client.events.list({ status: 'open', pageSize: 50 })

for await (const event of client.events.listAutoPaginate({ status: 'open' })) {
  console.log(event.name)
}

const event = await client.events.retrieve('evt_123')
const updated = await client.events.update('evt_123', { name: 'New name' })
```

Generate API keys in the 3Common organizer dashboard (`Settings → API Keys`). The key may also be supplied via the `THREECOMMON_API_KEY` environment variable.

## Configuration

```ts
new ThreeCommon({
  apiKey: '3co_…', // required (or via env var)
  baseUrl: 'https://api.3common.com', // default
  apiVersion: '2026-04-29', // pinned API version
  timeoutMs: 30_000, // per-request timeout
  maxRetries: 3, // automatic retries on 408/425/429/5xx
  retryDelay: { initialMs: 500, maxMs: 8000, jitter: true },
  fetch: customFetch, // override the fetch implementation
  logger: customLogger, // optional debug logger
  telemetry: true, // opt-out of anonymous telemetry
})
```

## Error handling

Every error thrown by the SDK is a subclass of `ThreeCommonError`. Branch with `instanceof`:

```ts
import {
  ThreeCommonNotFoundError,
  ThreeCommonRateLimitError,
  ThreeCommonAuthError,
} from '@3common/sdk'

try {
  await client.events.retrieve('evt_missing')
} catch (err) {
  if (err instanceof ThreeCommonNotFoundError) {
    // 404
  } else if (err instanceof ThreeCommonAuthError) {
    // 401 — bad or expired API key
  } else if (err instanceof ThreeCommonRateLimitError) {
    // 429 — err.retryAfterSeconds tells you when to retry
  } else {
    throw err
  }
}
```

Every error carries `code`, `message`, `httpStatus`, `requestId`, `details`, and `rawResponse`. `toString()` produces a single line including the request ID for log correlation:

```
[not_found] Event evt_missing not found (request_id=req-dfx-abc)
```

## Pagination

Two flavors:

```ts
// One page at a time
const { data, hasMore } = await client.events.list({ pageSize: 50 })

// All pages, lazy
for await (const event of client.events.listAutoPaginate()) {
  // ...
}
```

## Retries

Idempotent methods (GET, PATCH) retry automatically on `408`, `425`, `429`, `500`, `502`, `503`, `504` and on network errors. Backoff is exponential with full jitter, capped at `retryDelay.maxMs`. The SDK honors a server-provided `Retry-After` header on `429`.

`POST` and `DELETE` do not retry by default; pass an `Idempotency-Key` via per-request options to opt in (forward-compat — no v1 endpoints currently use this).

## Telemetry

The SDK sends a small, anonymized `Threecommon-Client-Telemetry` header on every request (SDK version, language, last-request latency). This helps debug performance reports from real customers without instrumenting their code. Disable globally:

```ts
const client = new ThreeCommon({ apiKey: '...', telemetry: false })
```

Or at runtime:

```ts
client.disableTelemetry()
```

The header never contains your API key, request bodies, or response bodies.

## Versioning

The SDK follows SemVer. The pinned **API version** (sent as `Threecommon-Version`) is independent — the API can evolve without breaking already-deployed SDKs. Bump `apiVersion` to opt into newer server behavior.

## Contributing

See the [repository CONTRIBUTING guide](https://github.com/3-Common/sdk/blob/main/CONTRIBUTING.md). Issues and PRs welcome.

## License

[MIT](./LICENSE)
