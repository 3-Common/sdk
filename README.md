# 3Common SDK

Official client libraries for the [3Common](https://3common.com) Public API.

| Language | Package | Status |
|----------|---------|--------|
| Node.js / TypeScript | [`@3-common/sdk`](https://www.npmjs.com/package/@3-common/sdk) | In development |
| Python | `threecommon` | Planned |
| Go | `github.com/3-Common/sdk/sdk-go` | In development |

The SDKs target **API v1**. The OpenAPI spec is published at `https://api.3common.com/docs/json` and a snapshot is committed at [`openapi/spec.yaml`](./openapi/spec.yaml).

## Install

Install snippets will appear here as each SDK ships its first release.

## Authentication

All SDKs accept an API key generated from the 3Common organizer dashboard (Settings → API Keys). The key is sent as a Bearer token; never expose it in browsers or commit it to source control.

## Documentation

- [API reference](https://api.3common.com/docs)
- Per-SDK documentation lives under each `sdk-*/` folder

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Bugs and feature requests via [GitHub Issues](https://github.com/3-Common/sdk/issues). Security reports: see [SECURITY.md](./SECURITY.md).

## License

[MIT](./LICENSE)
