import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  AutoChargeOutcome,
  AutoChargeResult,
  DeletedInvoice,
  Invoice,
  InvoiceCreateBody,
  InvoiceListParams,
  InvoicePaymentBody,
  InvoiceRefundBody,
  InvoiceRetrieveParams,
  InvoiceUpdateBody,
  InvoiceVoidBody,
  ListInvoicesResponse,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list — accepts the documented params and returns a typed ListInvoicesResponse.
expectType<Promise<ListInvoicesResponse>>(client.invoices.list({ status: 'open', pageSize: 50 }))
expectAssignable<InvoiceListParams>({ status: 'open', customerId: 'cnt_42' })
expectError<InvoiceListParams>({ status: 'not-a-status' })

// retrieve — id is a string; returns Invoice.
expectType<Promise<Invoice>>(client.invoices.retrieve('inv_123'))
expectAssignable<InvoiceRetrieveParams>({ fields: 'id,status' })

// create — body matches InvoiceCreateBody; returns Invoice.
declare const createBody: InvoiceCreateBody
expectType<Promise<Invoice>>(client.invoices.create(createBody))
expectAssignable<InvoiceCreateBody>({
  customerId: 'cnt_42',
  currency: 'USD',
  lineItems: [{ description: 'Consulting', quantity: 1, unitAmount: 50_000 }],
})
expectError<InvoiceCreateBody>({
  customerId: 'cnt_42',
  currency: 'EUR',
  lineItems: [],
})

// update — partial; returns Invoice.
declare const updateBody: InvoiceUpdateBody
expectType<Promise<Invoice>>(client.invoices.update('inv_123', updateBody))

// finalize — id only; returns Invoice.
expectType<Promise<Invoice>>(client.invoices.finalize('inv_123'))

// void — body is optional; returns Invoice.
expectType<Promise<Invoice>>(client.invoices.void('inv_123'))
declare const voidBody: InvoiceVoidBody
expectType<Promise<Invoice>>(client.invoices.void('inv_123', voidBody))

// recordPayment — body required with payment field; returns Invoice.
declare const paymentBody: InvoicePaymentBody
expectType<Promise<Invoice>>(client.invoices.recordPayment('inv_123', paymentBody))
expectAssignable<InvoicePaymentBody>({ payment: 50_000 })
expectAssignable<InvoicePaymentBody>({ payment: 50_000, idempotencyKey: 'pmt-1', note: 'wire' })

// autoCharge — id only; returns AutoChargeResult ({ invoice, outcome, failureCode? }).
expectType<Promise<AutoChargeResult>>(client.invoices.autoCharge('inv_123'))
declare const autoChargeResult: AutoChargeResult
expectType<Invoice>(autoChargeResult.invoice)
expectType<AutoChargeOutcome>(autoChargeResult.outcome)
expectAssignable<AutoChargeOutcome>('paid')
expectAssignable<AutoChargeOutcome>('failed')
expectError<AutoChargeOutcome>('refunded')

// refundPayment — id + paymentId + body; returns Invoice.
declare const refundBody: InvoiceRefundBody
expectType<Promise<Invoice>>(client.invoices.refundPayment('inv_123', 'pay_456', refundBody))
expectAssignable<InvoiceRefundBody>({ amount: 25_000 })
expectAssignable<InvoiceRefundBody>({
  amount: 25_000,
  reason: 'requested_by_customer',
  note: 'duplicate charge',
  idempotencyKey: 'rfnd-1',
})
// amount is required.
expectError<InvoiceRefundBody>({ reason: 'requested_by_customer' })

// deleteDraft — id only; returns DeletedInvoice ({ id }).
expectType<Promise<DeletedInvoice>>(client.invoices.deleteDraft('inv_123'))

// listAutoPaginate — returns AsyncIterableIterator<Invoice>.
expectAssignable<AsyncIterable<Invoice>>(client.invoices.listAutoPaginate({ status: 'open' }))
