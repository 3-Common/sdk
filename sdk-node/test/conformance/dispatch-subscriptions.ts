/**
 * Subscriptions-resource dispatcher for the conformance harness. Kept in
 * its own file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchSubscriptions(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.subscriptions.list(args)
    case 'retrieve': {
      const id = expectString(args['id'], 'retrieve')
      const params = args['params'] as Record<string, string> | undefined
      return client.subscriptions.retrieve(id, params)
    }
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.subscriptions.create(body)
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.subscriptions.update(id, body)
    }
    case 'activate':
      return client.subscriptions.activate(expectString(args['id'], 'activate'))
    case 'cancel': {
      const id = expectString(args['id'], 'cancel')
      const body = args['body'] as Record<string, unknown> | undefined
      return client.subscriptions.cancel(id, body)
    }
    case 'cancelImmediately': {
      const id = expectString(args['id'], 'cancelImmediately')
      const body = args['body'] as Record<string, unknown> | undefined
      return client.subscriptions.cancelImmediately(id, body)
    }
    case 'markUnpaid':
      return client.subscriptions.markUnpaid(expectString(args['id'], 'markUnpaid'))
    case 'bill':
      return client.subscriptions.bill(expectString(args['id'], 'bill'))
    case 'renew':
      return client.subscriptions.renew(expectString(args['id'], 'renew'))
    case 'previewUpcomingInvoice':
      return client.subscriptions.previewUpcomingInvoice(
        expectString(args['id'], 'previewUpcomingInvoice'),
      )
    case 'listAutoPaginate':
      return client.subscriptions.listAutoPaginate(args)
    case 'finalize':
    case 'void':
    case 'recordPayment':
    case 'autoCharge':
    case 'refundPayment':
    case 'count':
    case 'delete':
    case 'bulkUpsert':
    case 'listActivity':
    case 'listActivityAutoPaginate':
    case 'deleteDraft':
    case 'lookup':
    case 'grant':
    case 'consume':
    case 'archive':
    case 'unarchive':
      throw new Error(`subscriptions: unsupported method '${call.method}'`)
  }
}
