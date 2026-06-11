---
'@3common/sdk': minor
---

Add the `forms` resource. The new `client.forms` surface covers the full form
builder: `list`, `create`, `retrieve`, `update`, `duplicate`, element CRUD
(`addElement`, `updateElement`, `deleteElement`, `moveElement`), the "Other"
free-text toggle (`enableOtherOption`, `disableOtherOption`), conditional logic
rules (`addLogicRule`, `removeLogicRule`), and a `listAutoPaginate` iterator.
Includes typed `Form`, `FormElement`, `FormSummary`, the
`FormType`/`FormStatus`/`SubmitButtonAlign` unions, and per-endpoint request
body types.
