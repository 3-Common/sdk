/**
 * Permanently delete a draft invoice. Only drafts can be deleted — once an
 * invoice is finalized (it has a number), void it instead so the audit trail
 * stays intact.
 *
 * Run:
 *   npx tsx examples/invoices/delete-draft.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { id } = await client.invoices.deleteDraft('inv_replace_with_real_id')

console.log(`deleted draft invoice ${id}`)
