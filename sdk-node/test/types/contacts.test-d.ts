import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  AttachPaymentMethodResult,
  BulkUpsertContactsResult,
  Contact,
  ContactActivity,
  ContactActivityListParams,
  ContactBulkUpsertBody,
  ContactCountResult,
  ContactCreateBody,
  ContactListParams,
  ContactUpdateBody,
  ContactWithOrderDetails,
  DeletedContact,
  ListContactActivityResponse,
  ListContactsResponse,
  PaymentMethod,
  PaymentMethodSetupIntent,
  RemovedPaymentMethod,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts params and returns a typed ListContactsResponse.
expectType<Promise<ListContactsResponse>>(
  client.contacts.list({ filter: 'opted-in', pageSize: 50 }),
)
expectAssignable<ContactListParams>({ filter: 'opted-in', pageNumber: 0 })
expectError<ContactListParams>({ filter: 'not-a-status' })

// count — returns { count }.
expectType<Promise<ContactCountResult>>(client.contacts.count())

// retrieve — id is a string; returns Contact.
expectType<Promise<Contact>>(client.contacts.retrieve('cnt_123'))

// create — body matches ContactCreateBody; returns Contact.
declare const createBody: ContactCreateBody
expectType<Promise<Contact>>(client.contacts.create(createBody))
expectAssignable<ContactCreateBody>({ email: 'alex@example.com' })

// update — body has nested contact; returns the richer ContactWithOrderDetails.
declare const updateBody: ContactUpdateBody
expectType<Promise<ContactWithOrderDetails>>(client.contacts.update('cnt_123', updateBody))
expectAssignable<ContactUpdateBody>({
  contact: {
    firstName: 'Alex',
    lastName: 'Garcia',
    email: 'alex@example.com',
    status: 'opted-in',
  },
})
expectAssignable<ContactUpdateBody>({
  contact: {
    firstName: 'Alex',
    lastName: 'Garcia',
    email: 'alex@example.com',
    status: 'opted-in',
  },
  mergeWith: 'cnt_456',
  resolution: 'safe-merge',
})
expectError<ContactUpdateBody>({
  contact: {
    firstName: 'Alex',
    lastName: 'Garcia',
    email: 'alex@example.com',
    status: 'not-a-status',
  },
})

// delete — returns the DeletedContact echo.
expectType<Promise<DeletedContact>>(client.contacts.delete('cnt_123'))

// bulkUpsert — body matches ContactBulkUpsertBody; returns BulkUpsertContactsResult.
declare const bulkBody: ContactBulkUpsertBody
expectType<Promise<BulkUpsertContactsResult>>(client.contacts.bulkUpsert(bulkBody))

// listActivity — paginated activity feed.
expectType<Promise<ListContactActivityResponse>>(
  client.contacts.listActivity('cnt_123', { filter: 'checkout_session_completed' }),
)
expectAssignable<ContactActivityListParams>({ filter: 'email_sent', sort: 'oldest' })
expectError<ContactActivityListParams>({ sort: 'newest' })

// listAutoPaginate — returns AsyncIterableIterator<Contact>.
expectAssignable<AsyncIterable<Contact>>(client.contacts.listAutoPaginate({ filter: 'opted-in' }))

// listActivityAutoPaginate — returns AsyncIterableIterator<ContactActivity>.
expectAssignable<AsyncIterable<ContactActivity>>(
  client.contacts.listActivityAutoPaginate('cnt_123'),
)

// retrievePaymentMethod — returns the saved card or null.
expectType<Promise<PaymentMethod | null>>(client.contacts.retrievePaymentMethod('cnt_123'))

// attachPaymentMethod — body requires setupIntentId; returns the card + flag.
expectType<Promise<AttachPaymentMethodResult>>(
  client.contacts.attachPaymentMethod('cnt_123', { setupIntentId: 'seti_123' }),
)
expectError(client.contacts.attachPaymentMethod('cnt_123', {}))

// createPaymentMethodSetupIntent — id only; returns the setup intent.
expectType<Promise<PaymentMethodSetupIntent>>(
  client.contacts.createPaymentMethodSetupIntent('cnt_123'),
)

// removePaymentMethod — id + methodId; returns the removed flag.
expectType<Promise<RemovedPaymentMethod>>(client.contacts.removePaymentMethod('cnt_123', 'pm_456'))
