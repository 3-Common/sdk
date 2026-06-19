/**
 * Events-resource dispatcher for the conformance harness. Kept in its own
 * file so adding new resources doesn't bloat the shared runner.
 */

import type { ScenarioCall } from './scenario'
import type { ThreeCommon } from '@/client'

export function dispatchEvents(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.events.list(args)
    case 'retrieve': {
      const id = args['id']
      if (typeof id !== 'string') throw new Error('retrieve requires args.id (string)')
      const params = args['params'] as Record<string, string> | undefined
      return client.events.retrieve(id, params)
    }
    case 'update': {
      const id = args['id']
      if (typeof id !== 'string') throw new Error('update requires args.id (string)')
      const body = args['body'] as Record<string, unknown> | undefined
      if (body === undefined) throw new Error('update requires args.body')
      return client.events.update(id, body)
    }
    case 'listAutoPaginate':
      return client.events.listAutoPaginate(args)
    case 'create':
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
    case 'count':
    case 'delete':
    case 'bulkUpsert':
    case 'listActivity':
    case 'listActivityAutoPaginate':
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
      throw new Error(`events: unsupported method '${call.method}'`)
  }
}
