---
'@3common/sdk': minor
---

Add the `properties` resource. The new `client.properties` surface covers the
custom-property catalog: `list`, `retrieve`, `create`, `update`, and a
`listAutoPaginate` iterator. Includes the typed `Property` discriminated union
(keyed on `type`, with `options` on the `Select One` / `Select Multiple`
variants) plus the `PropertyType`, `PropertyObjectType`, `PropertyStatus`, and
`PropertyOption` aliases. `type` and `objectType` are fixed at creation;
properties are archived rather than deleted.
