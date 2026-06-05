import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  Entitlement,
  EntitlementConsumeBody,
  EntitlementGrantBody,
  EntitlementListParams,
  EntitlementLookupParams,
  EntitlementRetrieveParams,
  ListEntitlementsResponse,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListEntitlementsResponse.
expectType<Promise<ListEntitlementsResponse>>(
  client.entitlements.list({ featureKey: 'api_calls', minBalance: 1, pageSize: 50 }),
)
expectAssignable<EntitlementListParams>({ contactId: 'cnt_7', featureKey: 'api_calls' })
expectError<EntitlementListParams>({ minBalance: 'not-a-number' })

// retrieve — id is a string; returns Entitlement.
expectType<Promise<Entitlement>>(client.entitlements.retrieve('ent_123'))
expectAssignable<EntitlementRetrieveParams>({ fields: 'id,balance' })

// lookup — contactId + featureKey are required; returns Entitlement.
expectType<Promise<Entitlement>>(
  client.entitlements.lookup({ contactId: 'cnt_7', featureKey: 'api_calls' }),
)
expectAssignable<EntitlementLookupParams>({
  contactId: 'cnt_7',
  featureKey: 'api_calls',
  fields: 'id,balance',
})
// lookup requires both contactId and featureKey.
expectError<EntitlementLookupParams>({ contactId: 'cnt_7' })

// grant — body matches EntitlementGrantBody; returns Entitlement.
declare const grantBody: EntitlementGrantBody
expectType<Promise<Entitlement>>(client.entitlements.grant(grantBody))
expectAssignable<EntitlementGrantBody>({
  contactId: 'cnt_7',
  featureKey: 'api_calls',
  amount: 100,
  grantId: 'grant_1',
})

// consume — body matches EntitlementConsumeBody; returns Entitlement.
declare const consumeBody: EntitlementConsumeBody
expectType<Promise<Entitlement>>(client.entitlements.consume(consumeBody))
expectAssignable<EntitlementConsumeBody>({
  contactId: 'cnt_7',
  featureKey: 'api_calls',
  amount: 1,
})

// listAutoPaginate — returns AsyncIterableIterator<Entitlement>.
expectAssignable<AsyncIterable<Entitlement>>(
  client.entitlements.listAutoPaginate({ featureKey: 'api_calls' }),
)
