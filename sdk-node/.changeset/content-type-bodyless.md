---
'@3common/sdk': patch
---

Stop sending `Content-Type: application/json` on requests without a body.
`DELETE` and the action-style `POST` endpoints (`deleteDraft`, `finalize`,
`autoCharge`, `archive`, `unarchive`, `activate`, `markUnpaid`, `bill`,
`renew`) send no body, so advertising a JSON body caused servers that enforce
`Content-Type` against an empty body to reject the request with HTTP 400
(`FST_ERR_CTP_EMPTY_JSON_BODY`). `buildHeaders()` now sets `Content-Type` only
when the request actually carries a body. Fixes #91, #92, #93, #94, #95, #96.
