/**
 * Walk every opted-in contact for the host with the auto-paginator. Pages
 * are fetched lazily — one HTTP call per page, only when the previous
 * page's buffer drains.
 *
 * Run:
 *   npx tsx examples/contacts/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let total = 0
let lastEmail = ''
for await (const contact of client.contacts.listAutoPaginate({ filter: 'opted-in' })) {
  total += 1
  lastEmail = contact.email
  if (total % 100 === 0) console.log(`...processed ${String(total)} contacts`)
}

console.log(`walked ${String(total)} opted-in contacts total (last: ${lastEmail})`)
