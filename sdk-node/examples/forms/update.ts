/**
 * Edit a form's top-level settings. Only the fields you provide change.
 *
 * Run:
 *   npx tsx examples/forms/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const updated = await client.forms.update('frm_replace_with_real_id', {
  name: 'Renamed survey',
  status: 'active',
})

console.log(`updated ${updated.id} -> ${updated.name} (${updated.status})`)
