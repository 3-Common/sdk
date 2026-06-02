/**
 * Bulk-upsert contacts (e.g. from a CSV import). Deduplicated server-side by
 * email; existing rows are updated rather than rejected.
 *
 * Run:
 *   npx tsx examples/contacts/bulk-upsert.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { affected } = await client.contacts.bulkUpsert({
  contacts: [
    { email: 'aiden@example.com', firstName: 'Aidan', lastName: 'Garvey' },
    { email: 'ethan@example.com', firstName: 'Ethan', lastName: 'Toews'},
    { email: 'jaden@example.com', firstName: 'Jaden', lastName: 'Martens' },
    { email: 'zuhao@example.com', firstName: 'Zuhao', lastName: 'Fang' },
  ],
})

console.log(`upserted ${String(affected)} contacts`)
