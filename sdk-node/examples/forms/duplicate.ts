/**
 * Duplicate an existing form, renaming the copy.
 *
 * Run:
 *   npx tsx examples/forms/duplicate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const copy = await client.forms.duplicate('frm_replace_with_real_id', {
  name: 'Customer survey (copy)',
})

console.log(`duplicated into ${copy.id} - ${copy.name} (${copy.status})`)
