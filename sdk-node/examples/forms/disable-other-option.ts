/**
 * Disable the free-text "Other" choice on a selection element.
 *
 * Run:
 *   npx tsx examples/forms/disable-other-option.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.disableOtherOption(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
)

console.log(`disabled "Other" on element ${String(element.id)} (${element.type})`)
