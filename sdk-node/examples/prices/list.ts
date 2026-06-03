/**
 * List a product's active prices.
 *
 * Run:
 *   npx tsx examples/prices/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.prices.list({
  productId: 'prod_replace_with_real_id',
  active: true,
  pageSize: 25,
})

console.log(`got ${String(result.data.length)} prices (hasMore=${String(result.hasMore)})`)
for (const price of result.data) {
  console.log(
    `${price.id ?? '?'} — ${price.type ?? '?'} — ${String(price.unitAmount ?? 0)} ${price.currency ?? ''}`,
  )
}
