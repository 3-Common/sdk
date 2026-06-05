---
'@3common/sdk': minor
---

Add the `entitlements` resource. The new `client.entitlements` surface covers
balance lookups and grant management: `list`, `retrieve`, `lookup` (by contact
and feature), `grant` (manual top-up), `consume` (debit balance), and a
`listAutoPaginate` iterator.
