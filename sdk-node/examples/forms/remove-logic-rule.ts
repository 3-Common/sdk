/**
 * Remove the logic rule that reveals a target element from a source element.
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
  'elm_source_id',
  'elm_target_id',
)

console.log(`removed logic rule from element ${String(element.id)} (${element.type})`)
