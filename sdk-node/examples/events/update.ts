/**
 * Update an event's name.
 *
 * Run:
 *   npx tsx examples/events/update.ts
 */

import { ThreeCommon } from '@3-common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const updated = await client.events.update('evt_123', { name: 'New name' })

console.log(updated)
