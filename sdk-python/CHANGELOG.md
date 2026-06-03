# Changelog

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [SemVer](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## 0.7.0

### Added

- Features resource. The new `client.features` surface covers the feature
  catalog: `list`, `resolve` (resolve a feature's live value for a customer),
  `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
  `list_auto_paginate` iterator. Both sync and async surfaces.
- New public types on `threecommon.features`: `Feature`, `FeatureType`,
  `ResolvedFeature`, the `ResolvedFeatureValue` discriminated union and its
  `ResolvedFeatureBoolean`/`ResolvedFeatureQuantity`/`ResolvedFeatureEnum`/
  `ResolvedFeatureDuration` members, `CreateBody`, `UpdateBody`, `ListParams`,
  `RetrieveParams`, `ResolveParams`, and the `ListFeaturesResponse` envelope.

## 0.6.0

### Added

- Prices resource. The new `client.prices` surface covers the price catalog:
  `list`, `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
  `list_auto_paginate` iterator. Both sync and async surfaces.
- New public types on `threecommon.prices`: `Price`, `PriceRecurring`,
  `PriceFeature` (the boolean/quantity/enum/duration grant union) and its
  `PriceFeatureBoolean`/`PriceFeatureQuantity`/`PriceFeatureEnum`/
  `PriceFeatureDuration` members, `CreateBody`, `UpdateBody`, `ListParams`,
  `RetrieveParams`, the `ListPricesResponse` envelope, and the
  `PriceType`/`PriceCurrency`/`PriceInterval` literal unions.

## 0.5.0

### Added

- Entitlements resource. The new `client.entitlements` surface covers balance
  lookups and grant management: `list`, `retrieve`, `lookup` (by contact +
  feature), `grant` (manual top-up, idempotent on `grant_id`), `consume`
  (debit balance), and `list_auto_paginate`. Both sync and async surfaces.
- New public types on `threecommon.entitlements`: `Entitlement`,
  `EntitlementGrant`, `EntitlementGrantSource`, `GrantBody`, `ConsumeBody`,
  `ListParams`, `RetrieveParams`, `LookupParams`, and the
  `ListEntitlementsResponse` envelope.

## 0.4.0

### Added

- Contacts resource. The new `client.contacts` surface covers the full
  contact lifecycle: `list`, `count`, `retrieve`, `create`, `update`
  (with optional `merge_with` + `resolution` for absorbing a second
  contact during an email change), `delete`, `bulk_upsert`,
  `list_activity`, and both `list_auto_paginate` +
  `list_activity_auto_paginate` iterators. Both sync and async surfaces.
- New public types on `threecommon.contacts`: `Contact`,
  `ContactWithOrderDetails`, `ContactActivity`, `ContactProperty`,
  `ContactUpdate`, `CreateBody`, `UpdateBody`, `BulkUpsertBody`,
  `BulkUpsertItem`, `ListParams`, `ActivityListParams`, plus result
  envelopes `ListContactsResponse`, `ListActivityResponse`, `CountResult`,
  `BulkUpsertResult`, `DeleteResult`, and the lifecycle / merge / activity
  literal unions.

## 0.3.0

### Added

- Invoices: auto_charge, refund_payment, delete_draft methods (sync + async).
- Invoices: subscription_id filter on list().
- Invoices: AutoChargeOutcome, AutoChargeResult, DeletedInvoice, RefundBody types.

## 0.2.0

### Added

- Invoice write operations completing parity with the public REST surface, on
  both the sync and async clients:
  - `auto_charge` — off-session charge the customer's saved card
    (`POST /v1/invoices/{id}/auto_charge`). A decline resolves with
    `outcome="failed"` and a `failure_code` rather than raising; only network /
    processor 5xx errors raise.
  - `refund_payment` — refund all or part of a recorded payment
    (`POST /v1/invoices/{id}/payments/{paymentId}/refunds`). Idempotent on
    `body.idempotency_key`.
  - `delete_draft` — permanently remove a draft invoice
    (`DELETE /v1/invoices/{id}`).
- New public types on `threecommon.invoices`: `AutoChargeResult`,
  `AutoChargeOutcome`, `RefundBody`, and `DeletedInvoice`.

### Fixed

- `InvoiceStatus` now includes `payment_failed` (the state set after a failed
  off-session auto-charge); it was previously missing.
- Invoice `ListParams` now accepts the `subscription_id` filter the API
  supports; it was previously missing.

## 0.1.0

### Added

- Subscriptions resource. The new `client.subscriptions` surface covers the
  full subscription lifecycle: `list`, `retrieve`, `create`, `update`
  (mid-cycle change with proration), `activate`, `cancel`,
  `cancel_immediately`, `mark_unpaid`, `bill`, `renew`,
  `preview_upcoming_invoice`, and `list_auto_paginate`. Types and typed
  errors match the events / invoices resources. Both sync and async surfaces.

## 0.0.0

### Added

- Initial scaffolding.
- `ThreeCommon` (sync) and `AsyncThreeCommon` (async) clients.
- Events resource: `list`, `retrieve`, `update`, `list_auto_paginate`.
- Invoices resource: `list`, `retrieve`, `create`, `update`, `finalize`, `void`,
  `record_payment`, `list_auto_paginate`. Both sync and async surfaces.
- Typed exception tree (`AuthError`, `NotFoundError`, `RateLimitError`, …).
- Conformance harness running shared YAML scenarios against both clients.
