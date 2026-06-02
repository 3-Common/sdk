/**
 * Permanently remove a contact.
 *
 * Run:
 *   npx tsx examples/contacts/delete.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { id } = await client.contacts.delete('cnt_replace_with_real_id')

console.log(`deleted ${id}`)
