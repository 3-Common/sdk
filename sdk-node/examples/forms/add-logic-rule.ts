/**
 * Add a conditional-logic rule: reveal another element when the first option
 * of a selection question is chosen.
 *
 * Run:
 *   npx tsx examples/forms/add-logic-rule.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const element = await client.forms.addLogicRule('frm_replace_with_real_id', 'elm_source_id', {
  revealedElementId: 'elm_target_id',
  condition: { optionIndices: [0], operator: 'any_of' },
})

console.log(`added logic rule to element ${String(element.id)} (${element.type})`)
