/**
 * Rename a property and clear its description. `type` and `objectType` cannot
 * be modified. To retire a property, set `status` to `archived` instead of
 * deleting it.
 *
 * Run:
 *   npx tsx examples/properties/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const property = await client.properties.update('prop_replace_with_real_id', {
  name: 'Shirt size',
  description: null,
})

console.log(`updated ${property.id} - now named "${property.name}"`)
