---
'@3common/sdk': minor
---

Add the `prices` resource. The new `client.prices` surface covers the price
catalog: `list`, `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
`listAutoPaginate` iterator. Includes typed `Price`, `PriceFeature` (the
boolean/quantity/enum/duration grant union), `PriceRecurring`, and the
`PriceType`/`PriceCurrency`/`PriceInterval` unions.
