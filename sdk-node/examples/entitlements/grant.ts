/**
 * Manually grant entitlement units to a customer — admin top-ups, comp
 * credits, or migration. Idempotent on `grantId`: replaying the same id
 * returns the existing record without double-crediting.
 *
 * Run:
 *   npx tsx examples/entitlements/grant.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const ent = await client.entitlements.grant({
  contactId: 'cnt_replace_with_real_id',
  featureKey: 'api_calls',
  amount: 100,
  grantId: 'grant_2026_q2_goodwill',
  metadata: { reason: 'service-credit', approvedBy: 'ops' },
})

console.log(
  `granted — ${ent.contactId ?? '?'} now has ${String(ent.balance ?? 0)} ${ent.featureKey ?? '?'}`,
)
