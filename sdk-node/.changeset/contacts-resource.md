---
'@3common/sdk': minor
---

Add the `contacts` resource. The new `client.contacts` surface covers the
full contact lifecycle: `list`, `count`, `retrieve`, `create`, `update`
(with optional merge-on-conflict), `delete`, `bulkUpsert`, `listActivity`,
and both `listAutoPaginate` + `listActivityAutoPaginate` iterators.
