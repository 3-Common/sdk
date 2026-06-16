import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  ListPropertiesResponse,
  Property,
  PropertyCreateBody,
  PropertyListParams,
  PropertyUpdateBody,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list - accepts the documented params and returns a typed ListPropertiesResponse.
expectType<Promise<ListPropertiesResponse>>(
  client.properties.list({ objectType: 'contact', status: 'active', pageSize: 50 }),
)
expectAssignable<PropertyListParams>({ objectType: 'contact', propertyType: 'Select One' })
expectError<PropertyListParams>({ objectType: 'not-an-object-type' })
expectError<PropertyListParams>({ status: 'not-a-status' })

// retrieve - id is a string; returns Property.
expectType<Promise<Property>>(client.properties.retrieve('prop_123'))

// create - body matches PropertyCreateBody; returns Property.
declare const createBody: PropertyCreateBody
expectType<Promise<Property>>(client.properties.create(createBody))
expectAssignable<PropertyCreateBody>({
  type: 'Text',
  name: 'Dietary notes',
  status: 'active',
  objectType: 'contact',
})
expectAssignable<PropertyCreateBody>({
  type: 'Select One',
  name: 'T-shirt size',
  status: 'active',
  objectType: 'contact',
  options: [{ value: 's', label: 'Small' }],
})

// update - partial; description clears with null; returns Property.
declare const updateBody: PropertyUpdateBody
expectType<Promise<Property>>(client.properties.update('prop_123', updateBody))
expectAssignable<PropertyUpdateBody>({ name: 'Shirt size', description: null })

// listAutoPaginate - returns AsyncIterableIterator<Property>.
expectAssignable<AsyncIterable<Property>>(
  client.properties.listAutoPaginate({ objectType: 'contact' }),
)
