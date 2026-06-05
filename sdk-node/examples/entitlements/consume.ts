/**
 * Debit units from a customer's entitlement balance — call this when the
 * customer uses the metered feature. Debits `one_time_addon` grants first,
 * then `manual`, then `subscription_recurring`. Throws
 * `ThreeCommonConflictError` if the balance is insufficient.
 *
 * Run:
 *   npx tsx examples/entitlements/consume.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const ent = await client.entitlements.consume({
  contactId: 'cnt_replace_with_real_id',
  featureKey: 'api_calls',
  amount: 1,
  reason: 'POST /v1/generate',
})

console.log(`consumed 1 — ${String(ent.balance ?? 0)} ${ent.featureKey ?? '?'} remaining`)
