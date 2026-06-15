/**
 * List a host's contact properties, filtered to active ones.
 *
 * Run:
 *   npx tsx examples/properties/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.properties.list({
  objectType: 'contact',
  status: 'active',
  pageSize: 25,
})

console.log(`got ${String(result.data.length)} properties (hasMore=${String(result.hasMore)})`)
for (const property of result.data) {
  console.log(`${property.id} - ${property.name} [${property.type}] (${property.objectType})`)
}
