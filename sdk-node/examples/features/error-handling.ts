/**
 * Demonstrate the typed error tree on the features surface. Each subclass
 * extends `ThreeCommonError`; branch with `instanceof`.
 *
 * Run:
 *   npx tsx examples/features/error-handling.ts
 */

import {
  ThreeCommon,
  ThreeCommonAuthError,
  ThreeCommonConflictError,
  ThreeCommonNotFoundError,
  ThreeCommonRateLimitError,
  ThreeCommonValidationError,
} from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

try {
  // A feature `key` is unique per host — recreating an existing key conflicts.
  await client.features.create({ key: 'api_calls', name: 'API calls', type: 'quantity' })
} catch (err) {
  if (err instanceof ThreeCommonConflictError) {
    console.warn('a feature with this key already exists')
  } else if (err instanceof ThreeCommonValidationError) {
    console.warn(`validation: ${err.message}`)
  } else if (err instanceof ThreeCommonNotFoundError) {
    console.warn('feature not found')
  } else if (err instanceof ThreeCommonAuthError) {
    console.warn('bad or expired API key')
  } else if (err instanceof ThreeCommonRateLimitError) {
    const wait = err.retryAfterSeconds ?? 30
    console.warn(`rate limited; retry in ${String(wait)}s`)
  } else {
    throw err
  }
}
