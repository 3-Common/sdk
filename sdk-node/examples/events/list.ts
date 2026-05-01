/**
 * List events.
 *
 * Run:
 *   npx tsx examples/events/list.ts
 */

import { ThreeCommon } from '@3-common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const events = await client.events.list({ status: 'open', pageSize: 10 })

console.log(events)
