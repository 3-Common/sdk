/**
 * Retrieve a single form by id, including its elements and layout.
 *
 * Run:
 *   npx tsx examples/forms/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const form = await client.forms.retrieve('frm_replace_with_real_id')

console.log(`${form.id} - ${form.name}`)
console.log(`  type:   ${form.type}`)
console.log(`  status: ${form.status}`)
