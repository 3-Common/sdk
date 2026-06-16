/**
 * Move an element to a new position within the form.
 *
 * Run:
 *   npx tsx examples/forms/move-element.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const form = await client.forms.moveElement(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
  { position: 2 },
)

console.log(`moved element; form ${form.id} now has the updated layout`)
