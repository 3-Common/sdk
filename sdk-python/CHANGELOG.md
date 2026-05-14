# Changelog

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [SemVer](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial scaffolding.
- `ThreeCommon` (sync) and `AsyncThreeCommon` (async) clients.
- Events resource: `list`, `retrieve`, `update`, `list_auto_paginate`.
- Typed exception tree (`AuthError`, `NotFoundError`, `RateLimitError`, …).
- Conformance harness running shared YAML scenarios against both clients.
