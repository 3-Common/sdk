/**
 * Resolve a feature's live value for a customer — walks active subscriptions →
 * prices → feature grants. For quantity features it also reports the current
 * entitlement balance.
 *
 * Run:
 *   npx tsx examples/features/resolve.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const resolved = await client.features.resolve({
  contactId: 'cnt_replace_with_real_id',
  featureKey: 'api_calls',
})

console.log(`feature ${resolved.feature.key} [${resolved.value.type}]`)
switch (resolved.value.type) {
  case 'boolean':
    console.log(`  enabled: ${String(resolved.value.enabled)}`)
    break
  case 'quantity': {
    const q = resolved.value.quantity === null ? 'unlimited' : String(resolved.value.quantity)
    console.log(`  quantity: ${q}`)
    if (resolved.value.balance !== undefined) {
      console.log(`  balance:  ${String(resolved.value.balance)}`)
    }
    break
  }
  case 'enum':
    console.log(`  value: ${resolved.value.enumValue ?? 'none'}`)
    break
  case 'duration': {
    const d =
      resolved.value.durationDays === null
        ? 'unlimited'
        : `${String(resolved.value.durationDays)} days`
    console.log(`  duration: ${d}`)
    break
  }
}
console.log(`  from subscriptions: ${resolved.contributingSubscriptionIds.join(', ') || 'none'}`)
