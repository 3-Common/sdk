/**
 * Walk every standalone form for the host with the auto-paginator. Pages are
 * fetched lazily - one HTTP call per page, only when the previous page's
 * buffer drains.
 *
 * Run:
 *   npx tsx examples/forms/auto-paginate.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

let total = 0
let lastName = ''
for await (const form of client.forms.listAutoPaginate({ type: 'standalone' })) {
  total += 1
  lastName = form.name
  if (total % 100 === 0) console.log(`...processed ${String(total)} forms`)
}

console.log(`walked ${String(total)} standalone forms total (last: ${lastName})`)
