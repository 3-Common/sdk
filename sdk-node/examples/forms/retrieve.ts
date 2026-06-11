/**
 * Retrieve a single form, including its rows and elements, by id.
 *
 * Run:
 *   npx tsx examples/forms/retrieve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const form = await client.forms.retrieve('frm_replace_with_real_id')

console.log(`${form.name} (${form.type})`)
console.log(`  status:   ${form.status}`)
console.log(`  elements: ${String(form.elements.length)}`)
console.log(`  rows:     ${String(form.rows.length)}`)
