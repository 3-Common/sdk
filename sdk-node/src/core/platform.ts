import { arch, platform, release } from 'node:os'
import { performance } from 'node:perf_hooks'
import process from 'node:process'

/**
 * Environment-specific helpers. v1 is Node-only — When ship a browser
 * or edge-runtime target, this file gets a sibling and a small interface; the
 * abstraction does not exist preemptively.
 *
 * @internal
 */

/** Monotonic high-resolution timestamp in milliseconds. */
export function nowMs(): number {
  return performance.now()
}

/** Suffix appended to the SDK's User-Agent header (runtime + OS info). */
export function userAgentSuffix(): string {
  return `Node/${process.version}; ${platform()}-${arch()}-${release()}`
}

/**
 * Resolve the `fetch` implementation. Falls back to `globalThis.fetch`
 * (stable on Node ≥ 20). Throws if neither is available.
 */
export function resolveFetch(override: typeof fetch | undefined): typeof fetch {
  const provided = override ?? globalThis.fetch
  if (typeof provided !== 'function') {
    throw new TypeError(
      'globalThis.fetch is not a function. Use Node >= 20, or pass a `fetch` override on the client config.',
    )
  }
  return provided
}
