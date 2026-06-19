/**
 * Start saving a card for a contact. Returns a Stripe SetupIntent clientSecret
 * to confirm client-side with Stripe Elements, after which you call
 * `attachPaymentMethod` with the returned setupIntentId.
 *
 * Run:
 *   npx tsx examples/contacts/create-payment-method-setup-intent.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const intent = await client.contacts.createPaymentMethodSetupIntent('cnt_replace_with_real_id')

console.log(`setupIntentId: ${intent.setupIntentId}`)
console.log(`clientSecret:  ${intent.clientSecret}`)
console.log(`customerId:    ${intent.customerId}`)
