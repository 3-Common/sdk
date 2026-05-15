# Changelog

All notable changes to `github.com/3-Common/sdk/sdk-go` are documented in this
file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release scaffolding. 
- Events resource: `List`, `Retrieve`, `Update`, `ListAutoPaginate`.
- Invoices resource: `List`, `Retrieve`, `Create`, `Update`, `Finalize`, `Void`,
  `RecordPayment`, `ListAutoPaginate`.
- Range-over-func iterator (`Iter.All()`) on Go 1.23+ via build-tagged file.
- Maintainer-only live smoke test under `scripts/livesmoke`.
