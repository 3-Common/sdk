/**
 * Look up the unique entitlement for a (contact, feature) pair — the common
 * "how much does this customer have left?" check. Throws
 * `ThreeCommonNotFoundError` if no record exists yet.
 *
 * Run:
 *   npx tsx examples/entitlements/lookup.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const ent = await client.entitlements.lookup({
  contactId: 'cnt_replace_with_real_id',
  featureKey: 'api_calls',
})

console.log(
  `${ent.contactId ?? '?'} has ${String(ent.balance ?? 0)} ${ent.featureKey ?? '?'} remaining`,
)
