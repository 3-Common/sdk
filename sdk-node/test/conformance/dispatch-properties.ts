/**
 * Properties-resource dispatcher for the conformance harness. Kept in its own
 * file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchProperties(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.properties.list(args)
    case 'retrieve':
      return client.properties.retrieve(expectString(args['id'], 'retrieve'))
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.properties.create(
        body as unknown as Parameters<typeof client.properties.create>[0],
      )
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.properties.update(id, body)
    }
    case 'listAutoPaginate':
      return client.properties.listAutoPaginate(args)
    case 'count':
    case 'delete':
    case 'bulkUpsert':
    case 'listActivity':
    case 'listActivityAutoPaginate':
    case 'finalize':
    case 'void':
    case 'recordPayment':
    case 'autoCharge':
    case 'refundPayment':
    case 'deleteDraft':
    case 'activate':
    case 'cancel':
    case 'cancelImmediately':
    case 'markUnpaid':
    case 'bill':
    case 'renew':
    case 'previewUpcomingInvoice':
    case 'retrieveManageUrl':
    case 'lookup':
    case 'grant':
    case 'consume':
    case 'archive':
    case 'unarchive':
    case 'resolve':
    case 'duplicate':
    case 'addElement':
    case 'updateElement':
    case 'deleteElement':
    case 'moveElement':
    case 'enableOtherOption':
    case 'disableOtherOption':
    case 'addLogicRule':
    case 'removeLogicRule':
      throw new Error(`properties: unsupported method '${call.method}'`)
  }
}
