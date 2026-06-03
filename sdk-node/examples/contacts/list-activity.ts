/**
 * Fetch the activity feed (checkouts, refunds, scans, emails, invoice
 * payments) for a single contact.
 *
 * Run:
 *   npx tsx examples/contacts/list-activity.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.contacts.listActivity('cnt_replace_with_real_id', {
  pageSize: 20,
})

console.log(`got ${String(result.data.length)} activity records`)
for (const event of result.data) {
  console.log(`  ${event.createdAt} — ${event.type}`)
}
