/**
 * Update a contact's profile fields. The PATCH body wraps the new values
 * under `contact` and optionally accepts `mergeWith` + `resolution` when
 * absorbing a second contact during an email change.
 *
 * Run:
 *   npx tsx examples/contacts/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const updated = await client.contacts.update('cnt_replace_with_real_id', {
  contact: {
    firstName: 'Alex',
    lastName: 'Garcia',
    email: 'a.garcia@example.com',
    status: 'opted-in',
  },
})

console.log(`updated ${updated._id} → ${updated.email} (${updated.status})`)
