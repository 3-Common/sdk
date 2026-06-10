/**
 * Update a feature's display fields. `key` and `type` are immutable — archive
 * and create a new feature to change them.
 *
 * Run:
 *   npx tsx examples/features/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const feature = await client.features.update('feat_replace_with_real_id', {
  name: 'API requests',
  description: 'Monthly API request quota',
})

console.log(`updated ${feature.id ?? '?'} — ${feature.name ?? '?'}`)
