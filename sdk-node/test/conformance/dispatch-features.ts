/**
 * Features-resource dispatcher for the conformance harness. Kept in its own
 * file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchFeatures(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.features.list(args)
    case 'resolve':
      return client.features.resolve(
        args as unknown as Parameters<typeof client.features.resolve>[0],
      )
    case 'retrieve': {
      const id = expectString(args['id'], 'retrieve')
      const params = args['params'] as Record<string, string> | undefined
      return client.features.retrieve(id, params)
    }
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.features.create(body as unknown as Parameters<typeof client.features.create>[0])
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.features.update(id, body)
    }
    case 'archive':
      return client.features.archive(expectString(args['id'], 'archive'))
    case 'unarchive':
      return client.features.unarchive(expectString(args['id'], 'unarchive'))
    case 'listAutoPaginate':
      return client.features.listAutoPaginate(args)
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
    case 'count':
    case 'delete':
    case 'bulkUpsert':
    case 'listActivity':
    case 'listActivityAutoPaginate':
    case 'lookup':
    case 'grant':
    case 'consume':
    case 'duplicate':
    case 'addElement':
    case 'updateElement':
    case 'deleteElement':
    case 'moveElement':
    case 'enableOtherOption':
    case 'disableOtherOption':
    case 'addLogicRule':
    case 'removeLogicRule':
      throw new Error(`features: unsupported method '${call.method}'`)
  }
}
