/**
 * Create a recurring price with a metered feature grant. The `quantity` grant
 * refills the customer's entitlement balance on each renewal.
 *
 * Run:
 *   npx tsx examples/prices/create.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const price = await client.prices.create({
  productId: 'prod_replace_with_real_id',
  type: 'recurring',
  currency: 'USD',
  unitAmount: 1500,
  recurring: { interval: 'month', intervalCount: 1 },
  features: [{ featureKey: 'api_calls', type: 'quantity', quantity: 1000, rolloverEnabled: false }],
  nickname: 'Pro monthly',
  metadata: { tier: 'pro' },
})

console.log(`created ${price.id ?? '?'} — ${String(price.unitAmount ?? 0)} ${price.currency ?? ''}`)
