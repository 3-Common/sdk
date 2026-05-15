# Changelog

### 0.0.0

### Added

- Initial Node.js SDK skeleton with full `events` resource (`list`, `retrieve`, `update`, `listAutoPaginate`).
- `invoices` resource (`list`, `retrieve`, `create`, `update`, `finalize`, `void`, `recordPayment`, `listAutoPaginate`).
- Typed error hierarchy: `ThreeCommonError` plus per-status subclasses.
- HTTP layer with automatic retries (exponential backoff + full jitter), `Retry-After` honoring, configurable timeout, request abort signals.
- Opt-out telemetry header (`Threecommon-Client-Telemetry`).

## 0.1.0

### Minor Changes

- b5ed6e6: Add the `invoices` resource. The new `client.invoices` surface covers the full
  invoice lifecycle: `list`, `retrieve`, `create` (draft), `update` (revise while in
  draft), `finalize`, `void`, `recordPayment`, and `listAutoPaginate`. Types,
  typed errors (`ThreeCommonNotFoundError`, `ThreeCommonConflictError`, …), and
  retry policy match the events resource.

All notable changes to `@3common/sdk` are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/).
