/**
 * Invoices-resource dispatcher for the conformance harness. Kept in its own
 * file so adding new resources doesn't bloat the shared runner.
 */

import { expectBody, expectString, type ScenarioCall } from './scenario'

import type { ThreeCommon } from '@/client'

export function dispatchInvoices(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const args = call.args
  switch (call.method) {
    case 'list':
      return client.invoices.list(args)
    case 'retrieve': {
      const id = expectString(args['id'], 'retrieve')
      const params = args['params'] as Record<string, string> | undefined
      return client.invoices.retrieve(id, params)
    }
    case 'create': {
      const body = expectBody(args['body'], 'create')
      return client.invoices.create(body as unknown as Parameters<typeof client.invoices.create>[0])
    }
    case 'update': {
      const id = expectString(args['id'], 'update')
      const body = expectBody(args['body'], 'update')
      return client.invoices.update(id, body)
    }
    case 'finalize':
      return client.invoices.finalize(expectString(args['id'], 'finalize'))
    case 'void': {
      const id = expectString(args['id'], 'void')
      const body = args['body'] as Record<string, unknown> | undefined
      return client.invoices.void(id, body)
    }
    case 'recordPayment': {
      const id = expectString(args['id'], 'recordPayment')
      const body = expectBody(args['body'], 'recordPayment')
      return client.invoices.recordPayment(
        id,
        body as unknown as Parameters<typeof client.invoices.recordPayment>[1],
      )
    }
    case 'autoCharge':
      return client.invoices.autoCharge(expectString(args['id'], 'autoCharge'))
    case 'refundPayment': {
      const id = expectString(args['id'], 'refundPayment')
      const paymentId = expectString(args['paymentId'], 'refundPayment')
      const body = expectBody(args['body'], 'refundPayment')
      return client.invoices.refundPayment(
        id,
        paymentId,
        body as unknown as Parameters<typeof client.invoices.refundPayment>[2],
      )
    }
    case 'deleteDraft':
      return client.invoices.deleteDraft(expectString(args['id'], 'deleteDraft'))
    case 'listAutoPaginate':
      return client.invoices.listAutoPaginate(args)
    case 'activate':
    case 'cancel':
    case 'cancelImmediately':
    case 'markUnpaid':
    case 'bill':
    case 'renew':
    case 'previewUpcomingInvoice':
      throw new Error(`invoices: unsupported method '${call.method}'`)
  }
}
