/**
 * Edit an existing element's prompt. Only the fields you provide change.
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
  },
)

console.log(`updated ${String(element.id)} (${element.type}): ${element.prompt}`)
