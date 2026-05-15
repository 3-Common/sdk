/**
 * List open invoices for a customer.
 *
 * Run:
 *   npx tsx examples/invoices/list.ts
 */

import { ThreeCommon } from '@3-common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.invoices.list({
  status: 'open',
  customerId: 'cnt_replace_with_real_id',
  pageSize: 25,
})

console.log(`got ${String(result.data.length)} invoices (hasMore=${String(result.hasMore)})`)
for (const inv of result.data) {
  console.log(`${inv.id ?? '?'} — ${inv.status ?? '?'} — due ${String(inv.amountDue ?? 0)}`)
}
