/**
 * Demonstrate the typed error tree on the entitlements surface. Each subclass
 * extends `ThreeCommonError`; branch with `instanceof`.
 *
 * Run:
 *   npx tsx examples/entitlements/error-handling.ts
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
  await client.entitlements.consume({
    contactId: 'cnt_replace_with_real_id',
    featureKey: 'api_calls',
    amount: 1_000_000,
  })
} catch (err) {
  if (err instanceof ThreeCommonConflictError) {
    console.warn('insufficient balance — top up before consuming')
  } else if (err instanceof ThreeCommonNotFoundError) {
    console.warn('no entitlement record for this contact + feature')
  } else if (err instanceof ThreeCommonValidationError) {
    console.warn(`validation: ${err.message}`)
  } else if (err instanceof ThreeCommonAuthError) {
    console.warn('bad or expired API key')
  } else if (err instanceof ThreeCommonRateLimitError) {
    const wait = err.retryAfterSeconds ?? 30
    console.warn(`rate limited; retry in ${String(wait)}s`)
  } else {
    throw err
  }
}
