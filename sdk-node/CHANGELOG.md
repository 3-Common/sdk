# Changelog

All notable changes to `@3-common/sdk` are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Initial Node.js SDK skeleton with full `events` resource (`list`, `retrieve`, `update`, `listAutoPaginate`).
- Typed error hierarchy: `ThreeCommonError` plus per-status subclasses.
- HTTP layer with automatic retries (exponential backoff + full jitter), `Retry-After` honoring, configurable timeout, request abort signals.
- Opt-out telemetry header (`Threecommon-Client-Telemetry`).
