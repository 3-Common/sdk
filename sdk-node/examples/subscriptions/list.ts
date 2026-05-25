/**
 * List active subscriptions for a customer.
 *
 * Run:
 *   npx tsx examples/subscriptions/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.subscriptions.list({
  status: 'active',
  contactId: 'cnt_replace_with_real_id',
  pageSize: 25,
})

console.log(`got ${String(result.data.length)} subscriptions (hasMore=${String(result.hasMore)})`)
for (const sub of result.data) {
  console.log(`${sub.id ?? '?'} — ${sub.status ?? '?'} — renews ${sub.currentPeriodEnd ?? '?'}`)
}
