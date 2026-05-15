/**
 * Create a draft invoice and finalize it. Finalizing assigns a sequential
 * number, stamps `issuedAt`, and transitions the invoice to `open`.
 *
 * Run:
 *   npx tsx examples/invoices/create-and-finalize.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const draft = await client.invoices.create({
  customerId: 'cnt_replace_with_real_id',
  currency: 'USD',
  lineItems: [
    { description: 'Consulting — May 2026', quantity: 8, unitAmount: 12_500 }, // 8 × $125
    { description: 'Onboarding fee', quantity: 1, unitAmount: 50_000 }, //         $500
  ],
  notes: 'Net 30. Wire transfer preferred.',
})

console.log(`drafted ${draft.id ?? '?'} — total ${String(draft.total ?? 0)} USD`)

const issued = await client.invoices.finalize(draft.id ?? '')

console.log(`finalized ${issued.id ?? '?'} as ${issued.number ?? '?'} [${issued.status ?? '?'}]`)
