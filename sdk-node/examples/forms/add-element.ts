/**
 * Add a required text question to a form.
 *
 * Run:
 *   npx tsx examples/forms/add-element.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.addElement('frm_replace_with_real_id', {
  prompt: 'What is your name?',
  type: 'Text',
  required: true,
})

console.log(`added element ${String(element.id)} (${element.type})`)
