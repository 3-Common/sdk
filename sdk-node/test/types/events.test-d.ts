import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  Event,
  EventListParams,
  EventRetrieveParams,
  EventUpdateBody,
  ListEventsResponse,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListEventsResponse.
expectType<Promise<ListEventsResponse>>(client.events.list({ status: 'open', pageSize: 50 }))
expectAssignable<EventListParams>({ status: 'open' })
expectError<EventListParams>({ status: 'not-a-status' })

// retrieve — id is a string; returns Event.
expectType<Promise<Event>>(client.events.retrieve('evt_123'))
expectAssignable<EventRetrieveParams>({ fields: 'id,name' })

// update — body matches EventUpdateBody; returns Event.
declare const body: EventUpdateBody
expectType<Promise<Event>>(client.events.update('evt_123', body))

// listAutoPaginate — returns AsyncIterableIterator<Event>.
expectAssignable<AsyncIterable<Event>>(client.events.listAutoPaginate({ status: 'open' }))
