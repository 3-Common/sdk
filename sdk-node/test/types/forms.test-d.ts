import { expectAssignable, expectError, expectType } from 'tsd'

import type {
  DeletedElement,
  Form,
  FormCreateBody,
  FormElement,
  FormListParams,
  FormSummary,
  FormUpdateBody,
  ListFormsResponse,
  ThreeCommon,
} from '@3common/sdk'

declare const client: ThreeCommon

// list - accepts the documented params and returns a typed ListFormsResponse.
expectType<Promise<ListFormsResponse>>(client.forms.list({ type: 'standalone', pageSize: 50 }))
expectAssignable<FormListParams>({ type: 'order', page: 1 })
// @ts-expect-error testing an invalid type intentionally
expectError<FormListParams>({ type: 'not-a-form-type' })

// retrieve - id is a string; returns Form.
expectType<Promise<Form>>(client.forms.retrieve('frm_123'))

// create - body matches FormCreateBody; returns Form.
declare const createBody: FormCreateBody
expectType<Promise<Form>>(client.forms.create(createBody))
expectAssignable<FormCreateBody>({ name: 'Registration', type: 'standalone' })

// update - partial settings; returns Form.
declare const updateBody: FormUpdateBody
expectType<Promise<Form>>(client.forms.update('frm_123', updateBody))
expectAssignable<FormUpdateBody>({ status: 'active', submitButtonText: 'Sign up' })

// duplicate - id + optional body; returns Form.
expectType<Promise<Form>>(client.forms.duplicate('frm_123', { name: 'Copy', status: 'draft' }))
expectType<Promise<Form>>(client.forms.duplicate('frm_123'))

// element CRUD.
expectType<Promise<FormElement>>(
  client.forms.addElement('frm_123', { prompt: 'Name?', type: 'Text', required: true }),
)
expectType<Promise<FormElement>>(
  client.forms.updateElement('frm_123', 'elm_1', { prompt: 'Full name?', required: false }),
)
expectType<Promise<DeletedElement>>(client.forms.deleteElement('frm_123', 'elm_1'))
expectType<Promise<Form>>(client.forms.moveElement('frm_123', 'elm_1', { position: 2 }))

// other-option toggles.
expectType<Promise<FormElement>>(
  client.forms.enableOtherOption('frm_123', 'elm_1', { otherPrompt: 'Other' }),
)
expectType<Promise<FormElement>>(client.forms.disableOtherOption('frm_123', 'elm_1'))

// logic rules.
expectType<Promise<FormElement>>(
  client.forms.addLogicRule('frm_123', 'elm_1', {
    revealedElementId: 'elm_2',
    condition: { optionIndices: [0], operator: 'any_of' },
  }),
)
expectType<Promise<FormElement>>(client.forms.removeLogicRule('frm_123', 'elm_1', 'elm_2'))

// listAutoPaginate - returns AsyncIterableIterator<FormSummary>.
expectAssignable<AsyncIterable<FormSummary>>(client.forms.listAutoPaginate({ type: 'standalone' }))
