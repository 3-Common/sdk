/**
 * Iterate every open invoice for a customer, transparently fetching each
 * page as the previous one drains.
 *
 * Run:
 *   npx tsx examples/invoices/auto-paginate.ts
 */

import { ThreeCommon } from '@3-common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let totalDue = 0
for await (const invoice of client.invoices.listAutoPaginate({
  status: 'open',
  customerId: 'cnt_replace_with_real_id',
})) {
  totalDue += invoice.amountDue ?? 0
}

console.log(`total amount due across all open invoices: ${String(totalDue)} cents`)
