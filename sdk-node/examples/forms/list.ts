/**
 * List the host's standalone forms.
 *
 * Run:
 *   npx tsx examples/forms/list.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const result = await client.forms.list({
  type: 'standalone',
  pageSize: 50,
})

console.log(`got ${String(result.data.length)} forms (hasMore=${String(result.hasMore)})`)
for (const form of result.data) {
  console.log(`${form.id} - ${form.name} (${form.status}, ${String(form.numElements)} elements)`)
}
