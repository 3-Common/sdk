"""Pinned 3Common Public API version.

Sent as the `Threecommon-Version` header on every request. The server uses
this to dispatch to the matching version of its internal handlers, so older
SDKs continue to receive the response shape they were compiled against even
after the API evolves.
"""

#: API version this SDK is built against.
API_VERSION = "2026-04-29"

#: Path segment appended to the configured base URL when constructing
#: request URLs. Pinned to /v1 for now; will become configurable when v2
#: ships.
API_PATH = "/v1"
