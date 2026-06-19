/**
 * Entitlements-resource dispatcher for the conformance harness. Kept in its
 * own file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchEntitlements(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.entitlements.list(args)
    case 'retrieve': {
      const id = expectString(args['id'], 'retrieve')
      const params = args['params'] as Record<string, string> | undefined
      return client.entitlements.retrieve(id, params)
    }
    case 'lookup':
      return client.entitlements.lookup(
        args as unknown as Parameters<typeof client.entitlements.lookup>[0],
      )
    case 'grant': {
      const body = expectBody(args['body'], 'grant')
      return client.entitlements.grant(
        body as unknown as Parameters<typeof client.entitlements.grant>[0],
      )
    }
    case 'consume': {
      const body = expectBody(args['body'], 'consume')
      return client.entitlements.consume(
        body as unknown as Parameters<typeof client.entitlements.consume>[0],
      )
    }
    case 'listAutoPaginate':
      return client.entitlements.listAutoPaginate(args)
    case 'update':
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
    case 'previewUpcomingInvoice':
    case 'retrieveManageUrl':
    case 'count':
    case 'delete':
    case 'bulkUpsert':
    case 'listActivity':
    case 'listActivityAutoPaginate':
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
    case 'retrievePaymentMethod':
    case 'attachPaymentMethod':
    case 'createPaymentMethodSetupIntent':
    case 'removePaymentMethod':
      throw new Error(`entitlements: unsupported method '${call.method}'`)
  }
}
