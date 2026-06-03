/**
 * Create a new contact. Fails with 409 if the email is already in use for
 * this host.
 *
 * Run:
 *   npx tsx examples/contacts/create.ts
 */

import { ThreeCommon, ThreeCommonConflictError } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

try {
  const created = await client.contacts.create({
    email: 'guest@example.com',
    firstName: 'Alex',
    lastName: 'Garcia',
  })
  console.log(`created ${created.id} <${created.email}>`)
} catch (err) {
  if (err instanceof ThreeCommonConflictError) {
    console.warn('contact with that email already exists for this host')
  } else {
    throw err
  }
}
