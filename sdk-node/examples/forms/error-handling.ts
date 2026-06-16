/**
 * Demonstrate the typed error tree on the forms surface. Each subclass
 * extends `ThreeCommonError`; branch with `instanceof`.
 *
 * Run:
 *   npx tsx examples/forms/error-handling.ts
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
  // Expect: "validation: Name cannot be blank"
  await client.forms.create({ name: '', type: 'standalone' })
} catch (err) {
  if (err instanceof ThreeCommonValidationError) {
    console.warn(`validation: ${err.message}`)
  } else if (err instanceof ThreeCommonNotFoundError) {
    console.warn('form or element not found')
  } else if (err instanceof ThreeCommonAuthError) {
    console.warn('bad or expired API key')
  } else if (err instanceof ThreeCommonRateLimitError) {
    const wait = err.retryAfterSeconds ?? 30
    console.warn(`rate limited; retry in ${String(wait)}s`)
  } else {
    throw err
  }
}
