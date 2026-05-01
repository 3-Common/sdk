/**
 * The version of the 3Common Public API this SDK is built against.
 *
 * Sent as the `Threecommon-Version` header on every request. The server uses
 * this to dispatch to the matching version of its internal handlers, so older
 * SDKs continue to receive the response shape they were compiled against even
 * after the API evolves.
 */
export const API_VERSION = '2026-04-29' as const

/**
 * The path segment the SDK appends to the configured `baseUrl`. Pinned to v1
 * for now; will become configurable when API v2 ships.
 */
export const API_PATH = '/v1' as const

/**
 * Type of {@link API_VERSION}.
 */
export type ApiVersion = typeof API_VERSION
