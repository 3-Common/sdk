# Changelog

All notable changes to `github.com/3-Common/sdk/sdk-go` are documented in this
file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## 0.6.0

### Added

- Prices resource. The new `api.Prices` surface covers the price catalog:
  `List`, `Retrieve`, `Create`, `Update`, `Archive`, `Unarchive`, and a
  `ListAutoPaginate` iterator. Available as `api.Prices` and via `prices.New`
  for standalone use.
- New public types in `resources/prices`: `Price`, `Recurring`, `Feature` (a
  tagged union over `FeatureType` — boolean/quantity/enum/duration — with a
  custom `MarshalJSON` that emits each variant's shape and preserves `null`
  for unlimited quantity/duration grants), `CreateParams`, `UpdateParams`,
  `ListParams`, `RetrieveParams`, the `ListResponse` envelope, and the
  `Type`/`Currency`/`Interval` enums.

## 0.5.0

### Added

- Entitlements resource. The new `api.Entitlements` surface covers balance
  lookups and grant management: `List`, `Retrieve`, `Lookup` (by contact +
  feature), `Grant` (manual top-up, idempotent on `GrantID`), `Consume`
  (debit balance), and `ListAutoPaginate`. Available as `api.Entitlements`
  and via `entitlements.New` for standalone use.
- New public types in `resources/entitlements`: `Entitlement`, `Grant`,
  `GrantSource`, `ListParams`, `RetrieveParams`, `LookupParams`,
  `GrantParams`, `ConsumeParams`, and the `ListResponse` envelope.

## 0.4.0

### Added

- Contacts resource. The new `api.Contacts` surface covers the full contact
  lifecycle: `List`, `Count`, `Retrieve`, `Create`, `Update` (with optional
  `MergeWith` + `Resolution` for absorbing a second contact during an email
  change), `Delete`, `BulkUpsert`, `ListActivity`, and both
  `ListAutoPaginate` + `ListActivityAutoPaginate` iterators.
- New public types in `resources/contacts`: `Contact`, `WithOrderDetails`,
  `Activity`, `Property`, `ContactUpdate`, `CreateParams`, `UpdateParams`,
  `BulkUpsertParams`, `BulkUpsertItem`, `ListParams`, `ActivityListParams`,
  `ListResponse`, `ListActivityResponse`, `CountResult`, `BulkUpsertResult`,
  `DeleteResult`, plus the `Status`, `MergeResolution`, `QuickFilter`, and
  `ActivityType` enums.
- `contacts.ListParams.FilterWith(*filters.SerializableFilter)` convenience
  for serializing the typed filter builder onto `Filters`, matching the
  events resource.

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
