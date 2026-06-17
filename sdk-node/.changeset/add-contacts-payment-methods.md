---
'@3common/sdk': minor
---

Add saved-card management to the `contacts` resource. The `client.contacts`
surface gains `retrievePaymentMethod` (the saved card on file, or `null`),
`attachPaymentMethod` (persist a card from a confirmed Stripe SetupIntent,
reporting whether an existing card was replaced),
`createPaymentMethodSetupIntent` (start a Stripe SetupIntent to confirm
client-side with Stripe Elements), and `removePaymentMethod` (detach the saved
card). Includes the typed `PaymentMethod`, `PaymentMethodSetupIntent`,
`AttachPaymentMethodResult`, `RemovedPaymentMethod`, `AttachPaymentMethodBody`,
and `PaymentMethodStatus` aliases.
