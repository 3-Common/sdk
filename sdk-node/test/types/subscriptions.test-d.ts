import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  BillSubscriptionResult,
  ListSubscriptionsResponse,
  RenewSubscriptionResult,
  Subscription,
  SubscriptionCancelBody,
  SubscriptionCancelImmediatelyBody,
  SubscriptionCreateBody,
  SubscriptionInvoicePreview,
  SubscriptionListParams,
  SubscriptionManageUrl,
  SubscriptionRetrieveParams,
  SubscriptionUpdateBody,
  ThreeCommon,
  UpdateSubscriptionResult,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListSubscriptionsResponse.
expectType<Promise<ListSubscriptionsResponse>>(
  client.subscriptions.list({ status: 'active', pageSize: 50 }),
)
expectAssignable<SubscriptionListParams>({ status: 'past_due', contactId: 'cnt_42' })
expectError<SubscriptionListParams>({ status: 'not-a-status' })

// retrieve — id is a string; returns Subscription.
expectType<Promise<Subscription>>(client.subscriptions.retrieve('sub_123'))
expectAssignable<SubscriptionRetrieveParams>({ fields: 'id,status' })

// create — body matches SubscriptionCreateBody; returns Subscription.
declare const createBody: SubscriptionCreateBody
expectType<Promise<Subscription>>(client.subscriptions.create(createBody))
expectAssignable<SubscriptionCreateBody>({ priceId: 'price_7', contactId: 'cnt_42' })
expectAssignable<SubscriptionCreateBody>({
  items: [{ priceId: 'price_7', quantity: 2 }],
  customerEmail: 'a@b.com',
  trialDays: 14,
})

// update — partial; returns UpdateSubscriptionResult.
declare const updateBody: SubscriptionUpdateBody
expectType<Promise<UpdateSubscriptionResult>>(client.subscriptions.update('sub_123', updateBody))

// retrieveManageUrl — id only; returns SubscriptionManageUrl ({ url }).
expectType<Promise<SubscriptionManageUrl>>(client.subscriptions.retrieveManageUrl('sub_123'))
expectAssignable<SubscriptionManageUrl>({ url: 'https://portal.3common.com/s/sub_123' })

// activate — id only; returns Subscription.
expectType<Promise<Subscription>>(client.subscriptions.activate('sub_123'))

// cancel — body optional; returns Subscription.
expectType<Promise<Subscription>>(client.subscriptions.cancel('sub_123'))
declare const cancelBody: SubscriptionCancelBody
expectType<Promise<Subscription>>(client.subscriptions.cancel('sub_123', cancelBody))

// cancelImmediately — body optional; returns Subscription.
declare const cancelImmediatelyBody: SubscriptionCancelImmediatelyBody
expectType<Promise<Subscription>>(
  client.subscriptions.cancelImmediately('sub_123', cancelImmediatelyBody),
)

// markUnpaid — id only; returns Subscription.
expectType<Promise<Subscription>>(client.subscriptions.markUnpaid('sub_123'))

// bill — id only; returns BillSubscriptionResult.
expectType<Promise<BillSubscriptionResult>>(client.subscriptions.bill('sub_123'))

// renew — id only; returns RenewSubscriptionResult.
expectType<Promise<RenewSubscriptionResult>>(client.subscriptions.renew('sub_123'))

// previewUpcomingInvoice — id only; returns SubscriptionInvoicePreview | null.
expectType<Promise<SubscriptionInvoicePreview | null>>(
  client.subscriptions.previewUpcomingInvoice('sub_123'),
)

// listAutoPaginate — returns AsyncIterableIterator<Subscription>.
expectAssignable<AsyncIterable<Subscription>>(
  client.subscriptions.listAutoPaginate({ status: 'active' }),
)
