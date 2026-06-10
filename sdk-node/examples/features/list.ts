/**
 * List the feature catalog, filtered by value type and active status.
 *
 * Run:
 *   npx tsx examples/features/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.features.list({ type: 'quantity', active: true, pageSize: 25 })

console.log(`got ${String(result.data.length)} features (hasMore=${String(result.hasMore)})`)
for (const feature of result.data) {
  console.log(`${feature.id ?? '?'} — ${feature.key ?? '?'} — ${feature.type ?? '?'}`)
}
