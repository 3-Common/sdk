/**
 * Retrieve a single entitlement record by id, including its grant history.
 *
 * Run:
 *   npx tsx examples/entitlements/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const ent = await client.entitlements.retrieve('ent_replace_with_real_id')

console.log(`entitlement ${ent.id ?? '?'} [${ent.featureKey ?? '?'}]`)
console.log(`  contact        ${ent.contactId ?? '?'}`)
console.log(`  balance        ${String(ent.balance ?? 0)}`)
console.log(`  totalGranted   ${String(ent.totalGranted ?? 0)}`)
console.log(`  totalConsumed  ${String(ent.totalConsumed ?? 0)}`)
for (const grant of ent.grants ?? []) {
  console.log(
    `  grant ${grant.id} [${grant.source}] ${String(grant.remaining)}/${String(grant.amount)} remaining`,
  )
}
