/**
 * Contacts-resource dispatcher for the conformance harness. Kept in its own
 * file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchContacts(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.contacts.list(args)
    case 'count':
      return client.contacts.count()
    case 'retrieve':
      return client.contacts.retrieve(expectString(args['id'], 'retrieve'))
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.contacts.create(body as unknown as Parameters<typeof client.contacts.create>[0])
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.contacts.update(
        id,
        body as unknown as Parameters<typeof client.contacts.update>[1],
      )
    }
    case 'delete':
      return client.contacts.delete(expectString(args['id'], 'delete'))
    case 'bulkUpsert': {
      const body = expectBody(args['body'], 'bulkUpsert')
      return client.contacts.bulkUpsert(
        body as unknown as Parameters<typeof client.contacts.bulkUpsert>[0],
      )
    }
    case 'listActivity': {
      const id = expectString(args['id'], 'listActivity')
      const params = args['params'] as Parameters<typeof client.contacts.listActivity>[1]
      return client.contacts.listActivity(id, params)
    }
    case 'listAutoPaginate':
      return client.contacts.listAutoPaginate(args)
    case 'listActivityAutoPaginate': {
      const id = expectString(args['id'], 'listActivityAutoPaginate')
      const params = args['params'] as Parameters<
        typeof client.contacts.listActivityAutoPaginate
      >[1]
      return client.contacts.listActivityAutoPaginate(id, params)
    }
    case 'retrievePaymentMethod':
      return client.contacts.retrievePaymentMethod(
        expectString(args['id'], 'retrievePaymentMethod'),
      )
    case 'attachPaymentMethod': {
      const id = expectString(args['id'], 'attachPaymentMethod')
      const body = expectBody(args['body'], 'attachPaymentMethod')
      return client.contacts.attachPaymentMethod(
        id,
        body as unknown as Parameters<typeof client.contacts.attachPaymentMethod>[1],
      )
    }
    case 'createPaymentMethodSetupIntent':
      return client.contacts.createPaymentMethodSetupIntent(
        expectString(args['id'], 'createPaymentMethodSetupIntent'),
      )
    case 'removePaymentMethod': {
      const id = expectString(args['id'], 'removePaymentMethod')
      const methodId = expectString(args['methodId'], 'removePaymentMethod')
      return client.contacts.removePaymentMethod(id, methodId)
    }
    case 'finalize':
    case 'void':
    case 'recordPayment':
    case 'autoCharge':
    case 'refundPayment':
    case 'deleteDraft':
    case 'activate':
    case 'cancel':
    case 'cancelImmediately':
    case 'compNextCycle':
    case 'uncompNextCycle':
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
      throw new Error(`contacts: unsupported method '${call.method}'`)
  }
}
