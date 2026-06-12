/**
 * Forms-resource dispatcher for the conformance harness. Kept in its own file
 * so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchForms(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.forms.list(args)
    case 'retrieve':
      return client.forms.retrieve(expectString(args['id'], 'retrieve'))
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.forms.create(body as unknown as Parameters<typeof client.forms.create>[0])
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.forms.update(id, body)
    }
    case 'duplicate': {
      const id = expectString(args['id'], 'duplicate')
      const body = args['body'] === undefined ? undefined : expectBody(args['body'], 'duplicate')
      return client.forms.duplicate(id, body)
    }
    case 'addElement': {
      const id = expectString(args['id'], 'addElement')
      const body = expectBody(args['body'], 'addElement')
      return client.forms.addElement(
        id,
        body as unknown as Parameters<typeof client.forms.addElement>[1],
      )
    }
    case 'updateElement': {
      const id = expectString(args['id'], 'updateElement')
      const elementId = expectString(args['elementId'], 'updateElement')
      const body = expectBody(args['body'], 'updateElement')
      return client.forms.updateElement(id, elementId, body)
    }
    case 'deleteElement': {
      const id = expectString(args['id'], 'deleteElement')
      const elementId = expectString(args['elementId'], 'deleteElement')
      return client.forms.deleteElement(id, elementId)
    }
    case 'moveElement': {
      const id = expectString(args['id'], 'moveElement')
      const elementId = expectString(args['elementId'], 'moveElement')
      const body = expectBody(args['body'], 'moveElement')
      return client.forms.moveElement(
        id,
        elementId,
        body as unknown as Parameters<typeof client.forms.moveElement>[2],
      )
    }
    case 'enableOtherOption': {
      const id = expectString(args['id'], 'enableOtherOption')
      const elementId = expectString(args['elementId'], 'enableOtherOption')
      const body = expectBody(args['body'], 'enableOtherOption')
      return client.forms.enableOtherOption(
        id,
        elementId,
        body as unknown as Parameters<typeof client.forms.enableOtherOption>[2],
      )
    }
    case 'disableOtherOption': {
      const id = expectString(args['id'], 'disableOtherOption')
      const elementId = expectString(args['elementId'], 'disableOtherOption')
      return client.forms.disableOtherOption(id, elementId)
    }
    case 'addLogicRule': {
      const id = expectString(args['id'], 'addLogicRule')
      const elementId = expectString(args['elementId'], 'addLogicRule')
      const body = expectBody(args['body'], 'addLogicRule')
      return client.forms.addLogicRule(
        id,
        elementId,
        body as unknown as Parameters<typeof client.forms.addLogicRule>[2],
      )
    }
    case 'removeLogicRule': {
      const id = expectString(args['id'], 'removeLogicRule')
      const elementId = expectString(args['elementId'], 'removeLogicRule')
      const targetElementId = expectString(args['targetElementId'], 'removeLogicRule')
      return client.forms.removeLogicRule(id, elementId, targetElementId)
    }
    case 'listAutoPaginate':
      return client.forms.listAutoPaginate(args)
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
    case 'archive':
    case 'unarchive':
    case 'resolve':
      throw new Error(`forms: unsupported method '${call.method}'`)
  }
}
