/**
 * Build a typed filter and pass it to `events.list`. The `filter` namespace is
 * shared across resources — every endpoint that accepts `filters` consumes the
 * same builder.
 *
 * Run:
 *   npx tsx examples/events/filters.ts
 */

import { filter, ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const f = filter.and(
  filter.field('status').isAnyOf(['open', 'closed']),
  filter.or(
    filter.field('ticketSum').isGreaterThan(10),
    filter.field('revenueCents').isGreaterThanOrEqualTo(10_000),
  ),
)

const events = await client.events.list({ filters: f.serialize(), pageSize: 10 })

console.log(events)
