/**
 * Remove the logic rule on an element that reveals a given target element.
 *
 * Run:
 *   npx tsx examples/forms/remove-logic-rule.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.removeLogicRule(
  'frm_replace_with_real_id',
  'elm_replace_with_real_id',
  'elm_followup',
)

console.log(`removed logic rule from ${String(element.id)} (${element.type})`)
