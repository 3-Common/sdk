/**
 * Demonstrate the typed error tree on the contacts surface. Each subclass
 * extends `ThreeCommonError`; branch with `instanceof`.
 *
 * Run:
 *   npx tsx examples/contacts/error-handling.ts
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
  await client.contacts.create({ email: 'guest@example.com' })
} catch (err) {
  if (err instanceof ThreeCommonConflictError) {
    console.warn('contact already exists for this email')
  } else if (err instanceof ThreeCommonValidationError) {
    console.warn(`validation: ${err.message}`)
  } else if (err instanceof ThreeCommonNotFoundError) {
    console.warn('host or scope not found')
  } else if (err instanceof ThreeCommonAuthError) {
    console.warn('bad or expired API key')
  } else if (err instanceof ThreeCommonRateLimitError) {
    const wait = err.retryAfterSeconds ?? 30
    console.warn(`rate limited; retry in ${String(wait)}s`)
  } else {
    throw err
  }
}
