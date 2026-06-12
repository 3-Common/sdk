/**
 * Create a new, empty standalone form. The `type` is fixed at creation time.
 *
 * Run:
 *   npx tsx examples/forms/create.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const form = await client.forms.create({
  name: 'Registration',
  type: 'standalone',
})

console.log(`created ${form.id} - ${form.name} (${form.status})`)
