/**
 * Reactivate a previously archived price. Idempotent.
 *
 * Run:
 *   npx tsx examples/prices/unarchive.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const price = await client.prices.unarchive('price_replace_with_real_id')

console.log(`unarchived ${price.id ?? '?'} — active=${String(price.active ?? false)}`)
