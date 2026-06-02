/**
 * Get the total contact count for the host.
 *
 * Run:
 *   npx tsx examples/contacts/count.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { count } = await client.contacts.count()

console.log(`host has ${String(count)} contacts`)
