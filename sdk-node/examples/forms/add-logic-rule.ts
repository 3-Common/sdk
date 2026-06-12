/**
 * Add conditional-logic rules. Selection questions reveal a target element
 * based on which options are chosen; Yes/No questions reveal a target element
 * based on the answer value.
 *
 * Run:
 *   npx tsx examples/forms/add-logic-rule.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

// Selection question: reveal the target when the first option is chosen.
const selectSource = await client.forms.addLogicRule('frm_replace_with_real_id', 'elm_select_id', {
  revealedElementId: 'elm_target_id',
  condition: { optionIndices: [0], operator: 'any_of' },
})

console.log(`added selection rule to element ${String(selectSource.id)} (${selectSource.type})`)

// Yes/No question: reveal the target when the respondent answers "yes".
const yesNoSource = await client.forms.addLogicRule('frm_replace_with_real_id', 'elm_yes_no_id', {
  revealedElementId: 'elm_other_target_id',
  condition: { selectionType: 'is', value: true },
})

console.log(`added Yes/No rule to element ${String(yesNoSource.id)} (${yesNoSource.type})`)
