/**
 * Email a customer their invoice. An `open` (or `payment_failed`) invoice gets
 * a payment-link email (Pay button + the invoice PDF); a `paid` invoice gets
 * its receipt. `send` is re-callable, so it doubles as a resend — `draft` and
 * `void` invoices are rejected with a `409`.
 *
 * You can also email as part of finalizing by passing `{ sendEmail: true }` to
 * `finalize`, which this example shows as the one-step alternative.
 *
 * Run:
 *   npx tsx examples/invoices/send.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

// One step: finalize and email the payment link in a single call.
const issued = await client.invoices.finalize('inv_replace_with_real_id', { sendEmail: true })
console.log(`finalized ${issued.id ?? '?'} as ${issued.number ?? '?'} [${issued.status ?? '?'}]`)

// Or send (or re-send) the email for an already-finalized invoice.
const sent = await client.invoices.send('inv_replace_with_real_id')
console.log(`sent invoice ${sent.id ?? '?'} to ${sent.customerEmail ?? 'the customer'}`)
