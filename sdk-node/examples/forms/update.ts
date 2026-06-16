/**
 * Update a form's settings (publish it and customize the submit button).
 *
 * Run:
 *   npx tsx examples/forms/update.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const form = await client.forms.update('frm_replace_with_real_id', {
  name: 'Updated Registration',
  status: 'active',
  submitButtonText: 'Sign up',
})

console.log(`updated ${form.id} - ${form.name} (${form.status})`)
