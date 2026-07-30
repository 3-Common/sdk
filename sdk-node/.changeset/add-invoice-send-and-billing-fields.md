---
'@3common/sdk': minor
---

Add `client.invoices.send(id)`, wrapping `POST /v1/invoices/{id}/send`, which
emails the customer their invoice, a payment-link email for an `open` or
`payment_failed` invoice, or a receipt for a `paid` one. It is re-callable
(doubling as a resend) and rejects `draft` and `void` invoices with a `409`.
`client.invoices.finalize(id, options, params)` now accepts an optional
`InvoiceFinalizeParams` with `sendEmail`, mapping to
`POST /v1/invoices/{id}/finalize?sendEmail=true`, so finalizing can email the
customer in one step.

Add `ThreeCommonPaymentRequiredError` for `402 Payment Required` responses,
which the create-invoice, create-subscription, create-form, create-property,
and newsletter-send endpoints can now return; previously a `402` surfaced as a
generic `ThreeCommonValidationError`.

Surface the new `billingEmail` field on `Contact` (accepted by
`contacts.create` and `contacts.update`) and the read-only `nextCycleDiscount`
object on `Subscription`, which describes a one-time discount staged for the
next renewal.
