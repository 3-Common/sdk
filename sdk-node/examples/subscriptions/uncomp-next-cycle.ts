/**
 * Remove a staged comp so the next renewal bills at full price again — the
 * inverse of `compNextCycle`. A no-op when no comp is pending, and allowed on
 * a subscription in any state.
 *
 * Run:
 *   npx tsx examples/subscriptions/uncomp-next-cycle.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const sub = await client.subscriptions.uncompNextCycle('sub_replace_with_real_id')

console.log(`subscription ${sub.id ?? '?'} [${sub.status ?? '?'}]`)
console.log(`  next renewal (${sub.currentPeriodEnd ?? '?'}) will bill at full price`)
