/**
 * Detach a saved card from Stripe and remove it from the contact.
 *
 * Run:
 *   npx tsx examples/contacts/remove-payment-method.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { removed } = await client.contacts.removePaymentMethod(
  'cnt_replace_with_real_id',
  'pm_replace_with_real_id',
)

console.log(`removed: ${String(removed)}`)
