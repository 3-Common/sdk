/**
 * Delete an element from a form.
 *
 * Run:
 *   npx tsx examples/forms/delete-element.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { deletedElementId } = await client.forms.deleteElement(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
)

console.log(`deleted ${deletedElementId}`)
