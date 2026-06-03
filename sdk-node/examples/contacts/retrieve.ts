/**
 * Retrieve a single contact by id.
 *
 * Run:
 *   npx tsx examples/contacts/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const contact = await client.contacts.retrieve('cnt_replace_with_real_id')

console.log(`${contact.fullName} <${contact.email}>`)
console.log(`  status:      ${contact.status}`)
console.log(`  orders:      ${String(contact.orderSum)}`)
console.log(`  gross:       ${String(contact.grossSum)}`)
console.log(`  vendorId:    ${contact.vendorId}`)
