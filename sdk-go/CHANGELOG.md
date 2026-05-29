# Changelog

All notable changes to `github.com/3-Common/sdk/sdk-go` are documented in this
file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## 0.3.0

### Added

- Invoice write operations completing parity with the public REST surface:
  - `AutoCharge` — off-session charge the customer's saved card
    (`POST /v1/invoices/{id}/auto_charge`). A decline returns a result with
    `Outcome` "failed" and a `FailureCode` rather than an error; only network /
    processor 5xx errors return an error.
  - `RefundPayment` — refund all or part of a recorded payment
    (`POST /v1/invoices/{id}/payments/{paymentId}/refunds`). Idempotent on
    `RefundParams.IdempotencyKey`.
  - `DeleteDraft` — permanently remove a draft invoice
    (`DELETE /v1/invoices/{id}`).
- New types `AutoChargeResult`, `AutoChargeOutcome`, `RefundParams`, and
  `DeleteDraftResult` on the `invoices` package.
- `scripts/livesmoke` exercises the not-found path for each new method.

### Changed

- `invoices.Status` adds `StatusPaymentFailed` (the state set after a failed
  off-session auto-charge).
- `invoices.ListParams` adds the `SubscriptionID` filter accepted by
  `GET /v1/invoices`.

## 0.1.0

### Added

- Subscriptions resource: `List`, `Retrieve`, `Create`, `Update` (mid-cycle
  change with proration), `Activate`, `Cancel`, `CancelImmediately`,
  `MarkUnpaid`, `Bill`, `Renew`, `PreviewUpcomingInvoice`, `ListAutoPaginate`.
  Available as `api.Subscriptions` and via `subscriptions.New` for standalone
  use. Types and typed errors match the events / invoices resources.
- `scripts/livesmoke` exercises the subscriptions happy path and 404 path when
  `SMOKE_SUBSCRIPTION_ID` is set.

## 0.0.0

### Added

- Initial release scaffolding. 
- Events resource: `List`, `Retrieve`, `Update`, `ListAutoPaginate`.
- Invoices resource: `List`, `Retrieve`, `Create`, `Update`, `Finalize`, `Void`,
  `RecordPayment`, `ListAutoPaginate`.
- Range-over-func iterator (`Iter.All()`) on Go 1.23+ via build-tagged file.
- Maintainer-only live smoke test under `scripts/livesmoke`.
