/**
 * Soft-archive a feature. Idempotent.
 *
 * Run:
 *   npx tsx examples/features/archive.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const feature = await client.features.archive('feat_replace_with_real_id')

console.log(`archived ${feature.id ?? '?'} — active=${String(feature.active ?? false)}`)
