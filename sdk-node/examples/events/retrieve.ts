/**
 * Retrieve a single event by ID.
 *
 * Run:
 *   npx tsx examples/events/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const event = await client.events.retrieve('evt_123', { fields: 'id,name,start,status' })

console.log(event)
