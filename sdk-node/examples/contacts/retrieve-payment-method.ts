/**
 * Retrieve the saved card on file for a contact (or null when none is saved).
 *
 * Run:
 *   npx tsx examples/contacts/retrieve-payment-method.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const method = await client.contacts.retrievePaymentMethod('cnt_replace_with_real_id')

if (method === null) {
  console.log('no card on file')
} else {
  console.log(`${method.card.brand} ••••${method.card.last4}`)
  console.log(`  expires:  ${String(method.card.expMonth)}/${String(method.card.expYear)}`)
  console.log(`  status:   ${method.status}`)
}
