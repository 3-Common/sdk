/**
 * Soft-archive a price. Existing subscriptions keep billing; new subscriptions
 * can no longer select it until unarchived. Idempotent.
 *
 * Run:
 *   npx tsx examples/prices/archive.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const price = await client.prices.archive('price_replace_with_real_id')

console.log(`archived ${price.id ?? '?'} — active=${String(price.active ?? false)}`)
