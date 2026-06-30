# Changelog

## 0.11.0

### Minor Changes

- 4d31f5d: Add saved-card management to the `contacts` resource. The `client.contacts`
surface gains `retrievePaymentMethod` (the saved card on file, or `null`),
`attachPaymentMethod` (persist a card from a confirmed Stripe SetupIntent,
reporting whether an existing card was replaced),
`createPaymentMethodSetupIntent` (start a Stripe SetupIntent to confirm
client-side with Stripe Elements), and `removePaymentMethod` (detach the saved
card). Includes the typed `PaymentMethod`, `PaymentMethodSetupIntent`,
`AttachPaymentMethodResult`, `RemovedPaymentMethod`, `AttachPaymentMethodBody`,
and `PaymentMethodStatus` aliases.

## 0.10.0

### Minor Changes

- e27c8d4: Add `client.subscriptions.retrieveManageUrl(id)`, wrapping
`GET /v1/subscriptions/{id}/manage-url`. It returns the signed, customer-facing
self-service portal link (`{ url }`, typed as the new `SubscriptionManageUrl`)
scoped to a single subscription, which the subscriber can use to view, cancel,
or resume it.

## 0.9.0

### Minor Changes

- 9579d31: Add the `properties` resource. The new `client.properties` surface covers the
  custom-property catalog: `list`, `retrieve`, `create`, `update`, and a
  `listAutoPaginate` iterator. Includes the typed `Property` discriminated union
  (keyed on `type`, with `options` on the `Select One` / `Select Multiple`
  variants) plus the `PropertyType`, `PropertyObjectType`, `PropertyStatus`, and
  `PropertyOption` aliases. `type` and `objectType` are fixed at creation;
  properties are archived rather than deleted.

## 0.8.0

### Minor Changes

- 448e447: Add the `forms` resource. The new `client.forms` surface covers form authoring
  end to end: `list`, `retrieve`, `create`, `update`, `duplicate`, element CRUD
  (`addElement`, `updateElement`, `deleteElement`, `moveElement`), the
  `enableOtherOption`/`disableOtherOption` toggles for selection questions,
  conditional logic via `addLogicRule`/`removeLogicRule`, and a
  `listAutoPaginate` iterator. Includes typed `Form`, `FormSummary`,
  `FormElement`, `DeletedElement`, the `FormStatus`/`FormType` unions, and the
  request-body types for every endpoint.

## 0.7.1

### Patch Changes

- af89507: Stop sending `Content-Type: application/json` on requests without a body.
  `DELETE` and the action-style `POST` endpoints (`deleteDraft`, `finalize`,
  `autoCharge`, `archive`, `unarchive`, `activate`, `markUnpaid`, `bill`,
  `renew`) send no body, so advertising a JSON body caused servers that enforce
  `Content-Type` against an empty body to reject the request with HTTP 400
  (`FST_ERR_CTP_EMPTY_JSON_BODY`). `buildHeaders()` now sets `Content-Type` only
  when the request actually carries a body. Fixes #91, #92, #93, #94, #95, #96.

## 0.7.0

### Minor Changes

- 2dce39e: Add the `features` resource. The new `client.features` surface covers the
  feature catalog: `list`, `resolve` (resolve a feature's live value for a
  customer), `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
  `listAutoPaginate` iterator. Includes typed `Feature`, `FeatureType`
  (boolean/quantity/enum/duration), `ResolvedFeature`, and the
  `ResolvedFeatureValue` discriminated union.
- a658873: Add the `prices` resource. The new `client.prices` surface covers the price
  catalog: `list`, `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
  `listAutoPaginate` iterator. Includes typed `Price`, `PriceFeature` (the
  boolean/quantity/enum/duration grant union), `PriceRecurring`, and the
  `PriceType`/`PriceCurrency`/`PriceInterval` unions.

## 0.5.0

### Minor Changes

- 36878a7: Add the `entitlements` resource. The new `client.entitlements` surface covers
  balance lookups and grant management: `list`, `retrieve`, `lookup` (by contact
  and feature), `grant` (manual top-up), `consume` (debit balance), and a
  `listAutoPaginate` iterator.

## 0.4.0

### Minor Changes

- f942eb5: Add the `contacts` resource. The new `client.contacts` surface covers the
  full contact lifecycle: `list`, `count`, `retrieve`, `create`, `update`
  (with optional merge-on-conflict), `delete`, `bulkUpsert`, `listActivity`,
  and both `listAutoPaginate` + `listActivityAutoPaginate` iterators.

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
