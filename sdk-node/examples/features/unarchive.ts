/**
 * Reactivate a previously archived feature. Idempotent.
 *
 * Run:
 *   npx tsx examples/features/unarchive.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const feature = await client.features.unarchive('feat_replace_with_real_id')

console.log(`unarchived ${feature.id ?? '?'} — active=${String(feature.active ?? false)}`)
