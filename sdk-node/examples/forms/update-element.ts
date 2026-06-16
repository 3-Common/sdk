/**
 * Edit an existing element: change its prompt and make it optional.
 *
 * Run:
 *   npx tsx examples/forms/update-element.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.updateElement(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
  {
    prompt: 'What is your full name?',
    required: false,
  },
)

console.log(`updated element ${String(element.id)} (${element.type})`)
