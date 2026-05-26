# Changelog

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [SemVer](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## 0.1.0

### Added

- Subscriptions resource. The new `client.subscriptions` surface covers the
  full subscription lifecycle: `list`, `retrieve`, `create`, `update`
  (mid-cycle change with proration), `activate`, `cancel`,
  `cancel_immediately`, `mark_unpaid`, `bill`, `renew`,
  `preview_upcoming_invoice`, and `list_auto_paginate`. Types and typed
  errors match the events / invoices resources. Both sync and async surfaces.

## 0.0.0

### Added

- Initial scaffolding.
- `ThreeCommon` (sync) and `AsyncThreeCommon` (async) clients.
- Events resource: `list`, `retrieve`, `update`, `list_auto_paginate`.
- Invoices resource: `list`, `retrieve`, `create`, `update`, `finalize`, `void`,
  `record_payment`, `list_auto_paginate`. Both sync and async surfaces.
- Typed exception tree (`AuthError`, `NotFoundError`, `RateLimitError`, …).
- Conformance harness running shared YAML scenarios against both clients.
