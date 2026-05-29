# Changelog

## 0.3.0

### Minor Changes

- Add the remaining invoice write operations so the invoices resource reaches full parity with the public REST surface:
  - `client.invoices.autoCharge(id)` — off-session charge the customer's saved card (`POST /v1/invoices/{id}/auto_charge`). A decline resolves with `outcome: 'failed'` and a `failureCode` rather than throwing.
  - `client.invoices.refundPayment(id, paymentId, body)` — refund all or part of a recorded payment (`POST /v1/invoices/{id}/payments/{paymentId}/refunds`). Idempotent on `body.idempotencyKey`.
  - `client.invoices.deleteDraft(id)` — permanently remove a draft invoice (`DELETE /v1/invoices/{id}`).

  New exported types: `AutoChargeResult`, `AutoChargeOutcome`, `InvoiceRefundBody`, and `DeletedInvoice`.

  Also corrected two invoice types that had drifted from the API:
  - `InvoiceStatus` now includes `payment_failed` (the state set after a failed off-session auto-charge).
  - `InvoiceListParams` now exposes the `subscriptionId` filter accepted by `GET /v1/invoices`.

## 0.2.0

### Minor Changes

- Add the `subscriptions` resource. The new `client.subscriptions` surface covers the full subscription lifecycle: `list`, `retrieve`, `create`, `update`

## 0.2.0

### Minor Changes

- Add the `subscriptions` resource. The new `client.subscriptions` surface
  covers the full subscription lifecycle: `list`, `retrieve`, `create`,
  `update` (mid-cycle change with proration), `activate`, `cancel`,
  `cancelImmediately`, `markUnpaid`, `bill`, `renew`,
  `previewUpcomingInvoice`, and `listAutoPaginate`. Types and typed errors
  match the events / invoices resources.

## 0.1.0

### Minor Changes

- b5ed6e6: Add the `invoices` resource. The new `client.invoices` surface covers the full
  invoice lifecycle: `list`, `retrieve`, `create` (draft), `update` (revise while in
  draft), `finalize`, `void`, `recordPayment`, and `listAutoPaginate`. Types,
  typed errors (`ThreeCommonNotFoundError`, `ThreeCommonConflictError`, …), and
  retry policy match the events resource.

All notable changes to `@3common/sdk` are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/).

### 0.0.0

### Added

- Initial Node.js SDK skeleton with full `events` resource (`list`, `retrieve`, `update`, `listAutoPaginate`).
- `invoices` resource (`list`, `retrieve`, `create`, `update`, `finalize`, `void`, `recordPayment`, `listAutoPaginate`).
- Typed error hierarchy: `ThreeCommonError` plus per-status subclasses.
- HTTP layer with automatic retries (exponential backoff + full jitter), `Retry-After` honoring, configurable timeout, request abort signals.
- Opt-out telemetry header (`Threecommon-Client-Telemetry`).
