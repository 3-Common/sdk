/**
 * Void an invoice. Permitted from `draft` or `open`. Paid invoices cannot be
 * voided — issue a credit note or refund the payment instead.
 *
 * Run:
 *   npx tsx examples/invoices/void.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const voided = await client.invoices.void('inv_replace_with_real_id', {
  reason: 'Sent to the wrong customer',
})

console.log(`invoice ${voided.id ?? '?'} status: ${voided.status ?? '?'}`)
