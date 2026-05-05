package threecommon

// APIVersion is the version of the 3Common Public API this SDK is built
// against. Sent as the Threecommon-Version header on every request. The server
// uses it to dispatch to the matching version of its internal handlers, so
// older SDKs continue to receive the response shape they were compiled against
// even after the API evolves.
const APIVersion = "2026-04-29"

// APIPath is the path segment appended to [Config.BaseURL] when constructing
// request URLs. Pinned to /v1 for now; will become configurable when API v2
// ships.
const APIPath = "/v1"
