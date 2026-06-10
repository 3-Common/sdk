import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  ListPricesResponse,
  Price,
  PriceCreateBody,
  PriceListParams,
  PriceRetrieveParams,
  PriceUpdateBody,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListPricesResponse.
expectType<Promise<ListPricesResponse>>(
  client.prices.list({ productId: 'prod_7', active: true, type: 'recurring', pageSize: 50 }),
)
expectAssignable<PriceListParams>({ productId: 'prod_7', active: false })
expectError<PriceListParams>({ type: 'not-a-cadence' })

// retrieve — id is a string; returns Price.
expectType<Promise<Price>>(client.prices.retrieve('price_123'))
expectAssignable<PriceRetrieveParams>({ fields: 'id,unitAmount' })

// create — body matches PriceCreateBody; returns Price.
declare const createBody: PriceCreateBody
expectType<Promise<Price>>(client.prices.create(createBody))
expectAssignable<PriceCreateBody>({
  productId: 'prod_7',
  type: 'recurring',
  currency: 'USD',
  unitAmount: 1500,
  recurring: { interval: 'month', intervalCount: 1 },
})
expectAssignable<PriceCreateBody>({
  productId: 'prod_7',
  type: 'one_time',
  currency: 'CAD',
  unitAmount: 999,
  features: [{ featureKey: 'api_calls', type: 'quantity', quantity: 1000, rolloverEnabled: false }],
})

// update — partial; nullable clears; returns Price.
declare const updateBody: PriceUpdateBody
expectType<Promise<Price>>(client.prices.update('price_123', updateBody))
expectAssignable<PriceUpdateBody>({ unitAmount: 1200, nickname: null })

// archive / unarchive — id only; returns Price.
expectType<Promise<Price>>(client.prices.archive('price_123'))
expectType<Promise<Price>>(client.prices.unarchive('price_123'))

// listAutoPaginate — returns AsyncIterableIterator<Price>.
expectAssignable<AsyncIterable<Price>>(client.prices.listAutoPaginate({ active: true }))
