import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  Feature,
  FeatureCreateBody,
  FeatureListParams,
  FeatureResolveParams,
  FeatureRetrieveParams,
  FeatureUpdateBody,
  ListFeaturesResponse,
  ResolvedFeature,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListFeaturesResponse.
expectType<Promise<ListFeaturesResponse>>(
  client.features.list({ type: 'quantity', active: true, pageSize: 50 }),
)
expectAssignable<FeatureListParams>({ type: 'enum', active: false })
expectError<FeatureListParams>({ type: 'not-a-type' })

// resolve — contactId + featureKey are required; returns ResolvedFeature.
expectType<Promise<ResolvedFeature>>(
  client.features.resolve({ contactId: 'cnt_7', featureKey: 'api_calls' }),
)
expectAssignable<FeatureResolveParams>({ contactId: 'cnt_7', featureKey: 'api_calls' })
expectError<FeatureResolveParams>({ contactId: 'cnt_7' })

// retrieve — id is a string; returns Feature.
expectType<Promise<Feature>>(client.features.retrieve('feat_123'))
expectAssignable<FeatureRetrieveParams>({ fields: 'id,key,type' })

// create — body matches FeatureCreateBody; returns Feature.
declare const createBody: FeatureCreateBody
expectType<Promise<Feature>>(client.features.create(createBody))
expectAssignable<FeatureCreateBody>({ key: 'api_calls', name: 'API calls', type: 'quantity' })
expectAssignable<FeatureCreateBody>({
  key: 'plan_tier',
  name: 'Plan tier',
  type: 'enum',
  enumValues: ['free', 'pro', 'enterprise'],
})

// update — partial; nullable clears; returns Feature.
declare const updateBody: FeatureUpdateBody
expectType<Promise<Feature>>(client.features.update('feat_123', updateBody))
expectAssignable<FeatureUpdateBody>({ name: 'API requests', description: null })

// archive / unarchive — id only; returns Feature.
expectType<Promise<Feature>>(client.features.archive('feat_123'))
expectType<Promise<Feature>>(client.features.unarchive('feat_123'))

// listAutoPaginate — returns AsyncIterableIterator<Feature>.
expectAssignable<AsyncIterable<Feature>>(client.features.listAutoPaginate({ active: true }))
