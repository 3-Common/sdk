/**
 * Build a typed filter for the contacts list. The `filter` namespace is
 * shared across resources — every endpoint that accepts `filters` consumes
 * the same builder.
 *
 * The simple `filter` enum (`opted-in`, `unknown`, ...) and the rich
 * `filters` builder can be combined; the server ANDs them.
 *
 * Run:
 *   npx tsx examples/contacts/filters.ts
 */

import { filter, ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

// High-value opted-in contacts whose most recent order is in 2026.
const f = filter.and(
  filter.field('status').isAnyOf(['opted-in']),
  filter.field('grossSum').isGreaterThan(100_000),
  filter.or(
    filter.field('orderSum').isGreaterThanOrEqualTo(5),
    filter.field('lastOrder').isAfter('2026-01-01T00:00:00.000Z'),
  ),
)

const result = await client.contacts.list({
  filters: f.serialize(),
  sortField: 'grossSum',
  sortDirection: 'desc',
  pageSize: 25,
})

console.log(`matched ${String(result.data.length)} contacts (hasMore=${String(result.hasMore)})`)
for (const c of result.data) {
  console.log(`  ${c.fullName} <${c.email}> — gross ${String(c.grossSum)}`)
}
