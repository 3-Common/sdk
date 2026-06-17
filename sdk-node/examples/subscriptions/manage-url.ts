/**
 * Fetch the signed self-service portal URL for a subscription. Share the
 * returned link with the subscriber so they can view, cancel, or resume it.
 *
 * Run:
 *   npx tsx examples/subscriptions/manage-url.ts
 */

import { ThreeCommon } from '@3common/sdk'

const client = new ThreeCommon({
  apiKey: '3co_your_api_key_here',
})

const { url } = await client.subscriptions.retrieveManageUrl('sub_replace_with_real_id')

console.log(`manage URL: ${url}`)
