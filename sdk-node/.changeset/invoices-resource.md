---
'@3-common/sdk': minor
---

Add the `invoices` resource. The new `client.invoices` surface covers the full
invoice lifecycle: `list`, `retrieve`, `create` (draft), `update` (revise while in
draft), `finalize`, `void`, `recordPayment`, and `listAutoPaginate`. Types,
typed errors (`ThreeCommonNotFoundError`, `ThreeCommonConflictError`, …), and
retry policy match the events resource.
