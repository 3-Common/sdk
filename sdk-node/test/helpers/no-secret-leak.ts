import { expect } from 'vitest'

/**
 * Assert that a string never contains the API key. Use on every error message,
 * log line, telemetry payload, etc., that might be exposed to logs.
 */
export function assertNoSecretLeak(value: string, apiKey: string): void {
  expect(value, 'value should not contain the API key').not.toContain(apiKey)
}

/**
 * Assert that an arbitrary object's stringified form doesn't contain the
 * API key.
 */
export function assertNoSecretLeakInObject(value: unknown, apiKey: string): void {
  const json = typeof value === 'string' ? value : JSON.stringify(value)
  assertNoSecretLeak(json, apiKey)
}
