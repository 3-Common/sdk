---
'@3common/sdk': minor
---

Add the `features` resource. The new `client.features` surface covers the
feature catalog: `list`, `resolve` (resolve a feature's live value for a
customer), `retrieve`, `create`, `update`, `archive`, `unarchive`, and a
`listAutoPaginate` iterator. Includes typed `Feature`, `FeatureType`
(boolean/quantity/enum/duration), `ResolvedFeature`, and the
`ResolvedFeatureValue` discriminated union.
