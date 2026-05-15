/**
 * Retrieve a single invoice by ID, with line items and payments.
 *
 * Run:
 *   npx tsx examples/invoices/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const invoice = await client.invoices.retrieve('inv_replace_with_real_id')

console.log(`invoice ${invoice.id ?? '?'} [${invoice.status ?? '?'}]`)
console.log(`  total      ${String(invoice.total ?? 0)} ${invoice.currency ?? ''}`)
console.log(`  paid       ${String(invoice.amountPaid ?? 0)}`)
console.log(`  due        ${String(invoice.amountDue ?? 0)}`)
console.log(`  line items: ${String(invoice.lineItems?.length ?? 0)}`)
console.log(`  payments:   ${String(invoice.payments?.length ?? 0)}`)
