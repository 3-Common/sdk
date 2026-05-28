/**
 * Revise a draft invoice. Only legal while the invoice is in `draft` — once it
 * is finalized, void it and create a new one instead so the audit trail stays
 * intact. Only the fields you pass are changed; replacing `lineItems`
 * recomputes the totals server-side.
 *
 * The SDK method is `update()` (`PATCH /v1/invoices/{id}`); "revise" is the
 * domain term for editing a draft.
 *
 * Run:
 *   npx tsx examples/invoices/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const revised = await client.invoices.update('inv_replace_with_real_id', {
  notes: 'Net 30. Updated per customer request.',
  dueAt: '2026-07-01T00:00:00.000Z',
  lineItems: [
    { description: 'Consulting — May 2026', quantity: 10, unitAmount: 12_500 }, // bumped 8 → 10
  ],
})

console.log(`revised ${revised.id ?? '?'} [${revised.status ?? '?'}]`)
console.log(`  subtotal ${String(revised.subtotal ?? 0)}, total ${String(revised.total ?? 0)} USD`)
