/**
 * Attach a card to a contact from a confirmed Stripe SetupIntent. Replaces any
 * existing card on file.
 *
 * Run:
 *   npx tsx examples/contacts/attach-payment-method.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { data, replacedExisting } = await client.contacts.attachPaymentMethod(
  'cnt_replace_with_real_id',
  { setupIntentId: 'seti_replace_with_real_id' },
)

console.log(`saved ${data.card.brand} ••••${data.card.last4} (${data.id})`)
console.log(`  replaced existing: ${String(replacedExisting)}`)
