/**
 * Update a price's amount and nickname. To change type, currency, or product,
 * archive the price and create a new one instead.
 *
 * Run:
 *   npx tsx examples/prices/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const price = await client.prices.update('price_replace_with_real_id', {
  unitAmount: 1200,
  nickname: 'Pro monthly (promo)',
})

console.log(
  `updated ${price.id ?? '?'} — now ${String(price.unitAmount ?? 0)} ${price.currency ?? ''}`,
)
