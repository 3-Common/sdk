/**
 * List entitlement balance records, filtered by feature and a minimum
 * balance. Sorted by most-recently-updated.
 *
 * Run:
 *   npx tsx examples/entitlements/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.entitlements.list({
  featureKey: 'api_calls',
  minBalance: 1,
  pageSize: 25,
})

console.log(`got ${String(result.data.length)} entitlements (hasMore=${String(result.hasMore)})`)
for (const ent of result.data) {
  console.log(
    `${ent.id ?? '?'} — ${ent.contactId ?? '?'} — ${ent.featureKey ?? '?'} — balance ${String(ent.balance ?? 0)}`,
  )
}
