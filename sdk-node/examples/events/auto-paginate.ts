/**
 * Iterate every event matching the filter, paging automatically.
 *
 * Run:
 *   npx tsx examples/events/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

for await (const event of client.events.listAutoPaginate({ status: 'open' })) {
  console.log(event)
}
