/**
 * Retrieve a single property by ID. `Select One` and `Select Multiple`
 * properties additionally carry an `options` array.
 *
 * Run:
 *   npx tsx examples/properties/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const property = await client.properties.retrieve('prop_replace_with_real_id')

console.log(`property ${property.id} [${property.type}]`)
console.log(`  name     ${property.name}`)
console.log(`  object   ${property.objectType}`)
console.log(`  status   ${property.status}`)
if (property.type === 'Select One' || property.type === 'Select Multiple') {
  for (const option of property.options) {
    console.log(`  option   ${option.label} -> ${option.value}`)
  }
}
