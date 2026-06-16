/**
 * Demonstrate the typed error tree on the properties surface. Each subclass
 * extends `ThreeCommonError`; branch with `instanceof`.
 *
 * Run:
 *   npx tsx examples/properties/error-handling.ts
 */

import {
  ThreeCommon,
  ThreeCommonAuthError,
  ThreeCommonNotFoundError,
  ThreeCommonRateLimitError,
  ThreeCommonValidationError,
} from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

try {
  // `options` is required when type is `Select One` - omitting it triggers a
  // 400 validation error.
  await client.properties.create({
    type: 'Select One',
    name: 'T-shirt size',
    status: 'active',
    objectType: 'contact',
  })
} catch (err) {
  if (err instanceof ThreeCommonValidationError) {
    console.warn(`validation: ${err.message}`)
  } else if (err instanceof ThreeCommonNotFoundError) {
    console.warn('property not found')
  } else if (err instanceof ThreeCommonAuthError) {
    console.warn('bad or expired API key')
  } else if (err instanceof ThreeCommonRateLimitError) {
    const wait = err.retryAfterSeconds ?? 30
    console.warn(`rate limited; retry in ${String(wait)}s`)
  } else {
    throw err
  }
}
