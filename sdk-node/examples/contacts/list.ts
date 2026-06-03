/**
 * List opted-in contacts for the host.
 *
 * Run:
 *   npx tsx examples/contacts/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.contacts.list({
  filter: 'opted-in',
  pageSize: 50,
  sortField: 'mostRecentOrder',
  sortDirection: 'desc',
})

console.log(
  `got ${String(result.data.length)} contacts ` +
    `(hasMore=${String(result.hasMore)}, page=${String(result.pageNumber)})`,
)
for (const c of result.data) {
  console.log(`${c.id} — ${c.email} (${c.status})`)
}
