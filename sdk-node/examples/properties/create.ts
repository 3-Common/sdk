/**
 * Create a `Select One` property on contacts. `type` and `objectType` are
 * fixed at creation and cannot be changed later. `options` is required for
 * `Select One` and `Select Multiple` types.
 *
 * Run:
 *   npx tsx examples/properties/create.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const property = await client.properties.create({
  type: 'Select One',
  name: 'T-shirt size',
  description: 'Preferred shirt size for swag fulfillment.',
  status: 'active',
  objectType: 'contact',
  options: [
    { value: 's', label: 'Small' },
    { value: 'm', label: 'Medium' },
    { value: 'l', label: 'Large' },
  ],
})

console.log(`created ${property.id} - ${property.name} [${property.type}]`)
