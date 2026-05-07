/**
 * Demonstrate the SDK's typed error hierarchy. Every error thrown by the SDK
 * is a subclass of `ThreeCommonError`.
 *
 * Run:
 *   npx tsx examples/events/error-handling.ts
 */

import {
  ThreeCommon,
  ThreeCommonAuthError,
  ThreeCommonNotFoundError,
  ThreeCommonRateLimitError,
} from '@3-common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

try {
  const event = await client.events.retrieve('evt_definitely_does_not_exist_12345')
  console.log(event)
} catch (err) {
  if (err instanceof ThreeCommonNotFoundError) {
    console.log('Not found:', err.message, '(request_id:', err.requestId, ')')
  } else if (err instanceof ThreeCommonAuthError) {
    console.log('Bad or expired API key:', err.message)
  } else if (err instanceof ThreeCommonRateLimitError) {
    console.log('Rate limited; retry after', err.retryAfterSeconds, 'seconds')
  } else {
    throw err
  }
}
