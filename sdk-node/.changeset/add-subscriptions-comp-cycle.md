---
'@3common/sdk': minor
---

Add `client.subscriptions.compNextCycle(id)` and
`client.subscriptions.uncompNextCycle(id)`, wrapping
`POST /v1/subscriptions/{id}/comp-next-cycle` and
`POST /v1/subscriptions/{id}/uncomp-next-cycle`. `compNextCycle` stages a
one-time fully-free (100% off) next renewal cycle — consumed exactly once before
billing resumes at full price, and rejected on a `canceled` or `unpaid`
subscription. `uncompNextCycle` is the inverse: it clears a pending comp so the
next renewal bills at full price again (a no-op when none is staged). Both
return the updated `Subscription`.
