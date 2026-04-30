# Changelog

This file is the index across all language SDKs in this repository. Each SDK keeps its own detailed changelog under its directory.

| SDK | Changelog |
|-----|-----------|
| Node.js | [`sdk-node/CHANGELOG.md`](./sdk-node/CHANGELOG.md) |
| Python | [`sdk-python/CHANGELOG.md`](./sdk-python/CHANGELOG.md) |
| Go | [`sdk-go/CHANGELOG.md`](./sdk-go/CHANGELOG.md) |

Releases follow [Semantic Versioning](https://semver.org). Each SDK ships its own SemVer line; the API version they target is pinned via the `Threecommon-Version` header (`apiVersion` config field).

## Format

Each per-SDK changelog uses the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Sections per release: **Added**, **Changed**, **Deprecated**, **Removed**, **Fixed**, **Security**.

Commits follow [Conventional Commits](https://www.conventionalcommits.org/), and changelog entries are derived from those commits.
