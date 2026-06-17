---
'@3common/sdk': minor
---

Add `client.subscriptions.retrieveManageUrl(id)`, wrapping
`GET /v1/subscriptions/{id}/manage-url`. It returns the signed, customer-facing
self-service portal link (`{ url }`, typed as the new `SubscriptionManageUrl`)
scoped to a single subscription, which the subscriber can use to view, cancel,
or resume it.
