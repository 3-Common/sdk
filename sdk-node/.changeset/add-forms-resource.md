---
'@3common/sdk': minor
---

Add the `forms` resource. The new `client.forms` surface covers form authoring
end to end: `list`, `retrieve`, `create`, `update`, `duplicate`, element CRUD
(`addElement`, `updateElement`, `deleteElement`, `moveElement`), the
`enableOtherOption`/`disableOtherOption` toggles for selection questions,
conditional logic via `addLogicRule`/`removeLogicRule`, and a
`listAutoPaginate` iterator. Includes typed `Form`, `FormSummary`,
`FormElement`, `DeletedElement`, the `FormStatus`/`FormType` unions, and the
request-body types for every endpoint.
