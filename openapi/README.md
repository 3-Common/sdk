# OpenAPI

This folder holds the canonical OpenAPI spec snapshots used by every SDK in this repo.

## Files

| File | Purpose |
|------|---------|
| [`spec.json`](./spec.json) | Raw OpenAPI document, fetched from `https://api.3common.com/docs/json`. |
| [`spec.yaml`](./spec.yaml) | YAML rendering of the same spec. Easier to diff. |
| [`scripts/fetch.ts`](./scripts/fetch.ts) | Script that pulls the spec and writes both files. |

Nete: `spec.json` and `spec.yaml` both are regenerated together; do not edit either by hand.
## Sync flow

The [`openapi-sync.yml`](../.github/workflows/openapi-sync.yml) GitHub Action runs `fetch.ts` on a daily schedule. If the spec has changed, it opens a `chore(openapi): sync spec` PR. Reviewers verify the diff; merging the PR triggers per-SDK workflows that regenerate language-specific types.

## Running locally

```bash
cd openapi
npm install
OPENAPI_SOURCE_URL="https://api.3common.com/docs/json" npx tsx scripts/fetch.ts
```

The `OPENAPI_SOURCE_URL` variable lets you point at a non-production deployment when bringing up new endpoints.
