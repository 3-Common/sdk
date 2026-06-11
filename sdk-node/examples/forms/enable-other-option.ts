/**
 * Enable the "Other" free-text option on a selection element.
 *
 * Run:
 *   npx tsx examples/forms/enable-other-option.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.enableOtherOption(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
  {
    otherPrompt: 'Other (please specify)',
  },
)

console.log(`enabled "Other" on ${String(element.id)} (${element.type})`)
