import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  DeletedElement,
  Form,
  FormAddElementBody,
  FormAddLogicRuleBody,
  FormCreateBody,
  FormDuplicateBody,
  FormElement,
  FormEnableOtherOptionBody,
  FormListParams,
  FormMoveElementBody,
  FormSummary,
  FormUpdateBody,
  FormUpdateElementBody,
  ListFormsResponse,
} from './types'
import type { HttpClient } from '@/core/http-client'
import type { RequestOptions } from '@/types/public'

interface DetailEnvelope<T> {
  readonly data: T
}

/**
 * Forms service. Bound as `client.forms` on the main client.
 *
 * Wraps `GET /v1/forms`, `POST /v1/forms`, `GET /v1/forms/{formId}`,
 * `PATCH /v1/forms/{formId}`, `POST /v1/forms/{formId}/duplicate`,
 * `POST /v1/forms/{formId}/elements`,
 * `PATCH /v1/forms/{formId}/elements/{elementId}`,
 * `DELETE /v1/forms/{formId}/elements/{elementId}`,
 * `PUT /v1/forms/{formId}/elements/{elementId}/position`,
 * `PUT`/`DELETE /v1/forms/{formId}/elements/{elementId}/other-option`,
 * `POST /v1/forms/{formId}/elements/{elementId}/logic-rules`, and
 * `DELETE /v1/forms/{formId}/elements/{elementId}/logic-rules/{targetElementId}`.
 *
 * @public
 */
export interface FormsService {
  /**
   * List the authenticated host's forms.
   *
   * @example
   * ```ts
   * const { data, hasMore } = await client.forms.list({ type: 'standalone', pageSize: 50 })
   * ```
   */
  list(params?: FormListParams, options?: RequestOptions): Promise<ListFormsResponse>

  /**
   * Retrieve a single form by id, including all elements and layout.
   *
   * @example
   * ```ts
   * const form = await client.forms.retrieve('frm_123')
   * ```
   */
  retrieve(id: string, options?: RequestOptions): Promise<Form>

  /**
   * Create a new, empty form. `type` (`standalone` or `order`) is fixed at
   * creation time and cannot be changed afterwards.
   *
   * @example
   * ```ts
   * const form = await client.forms.create({ name: 'Registration', type: 'standalone' })
   * ```
   */
  create(body: FormCreateBody, options?: RequestOptions): Promise<Form>

  /**
   * Apply a partial update to a form's settings (name, status, submit-button
   * options, etc.). Elements are managed by the `*Element` methods.
   *
   * @example
   * ```ts
   * const form = await client.forms.update('frm_123', { status: 'active', submitButtonText: 'Sign up' })
   * ```
   */
  update(id: string, body: FormUpdateBody, options?: RequestOptions): Promise<Form>

  /**
   * Duplicate a form, copying its elements and layout into a new form.
   *
   * @example
   * ```ts
   * const copy = await client.forms.duplicate('frm_123', { name: 'Registration (Copy)', status: 'draft' })
   * ```
   */
  duplicate(id: string, body?: FormDuplicateBody, options?: RequestOptions): Promise<Form>

  /**
   * Append a new element (question or static element) to a form. Returns the
   * created element.
   *
   * @example
   * ```ts
   * const element = await client.forms.addElement('frm_123', {
   *   prompt: 'What is your name?',
   *   type: 'Text',
   *   required: true,
   * })
   * ```
   */
  addElement(id: string, body: FormAddElementBody, options?: RequestOptions): Promise<FormElement>

  /**
   * Apply a partial update to a single element. Returns the updated element.
   *
   * @example
   * ```ts
   * const element = await client.forms.updateElement('frm_123', 'elm_1', {
   *   prompt: 'What is your full name?',
   *   required: false,
   * })
   * ```
   */
  updateElement(
    id: string,
    elementId: string,
    body: FormUpdateElementBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Delete a single element from a form. Returns `{ deletedElementId }`.
   *
   * @example
   * ```ts
   * await client.forms.deleteElement('frm_123', 'elm_1')
   * ```
   */
  deleteElement(id: string, elementId: string, options?: RequestOptions): Promise<DeletedElement>

  /**
   * Move an element to a new position (and, for order forms, optionally a
   * different section). Returns the full updated form.
   *
   * @example
   * ```ts
   * const form = await client.forms.moveElement('frm_123', 'elm_1', { position: 2 })
   * ```
   */
  moveElement(
    id: string,
    elementId: string,
    body: FormMoveElementBody,
    options?: RequestOptions,
  ): Promise<Form>

  /**
   * Enable the free-text "Other" choice on a selection element. Returns the
   * updated element.
   *
   * @example
   * ```ts
   * const element = await client.forms.enableOtherOption('frm_123', 'elm_1', {
   *   otherPrompt: 'Other (please specify)',
   * })
   * ```
   */
  enableOtherOption(
    id: string,
    elementId: string,
    body: FormEnableOtherOptionBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Disable the free-text "Other" choice on a selection element. Returns the
   * updated element.
   *
   * @example
   * ```ts
   * const element = await client.forms.disableOtherOption('frm_123', 'elm_1')
   * ```
   */
  disableOtherOption(id: string, elementId: string, options?: RequestOptions): Promise<FormElement>

  /**
   * Add a conditional-logic rule to a selection or Yes/No element: when the
   * condition matches, the `revealedElementId` element is shown. Selection
   * questions take an `{ optionIndices, operator }` condition; Yes/No
   * questions take a `{ selectionType, value }` condition. Returns the
   * updated source element.
   *
   * @example
   * ```ts
   * // Selection question: reveal when the first option is chosen.
   * const element = await client.forms.addLogicRule('frm_123', 'elm_1', {
   *   revealedElementId: 'elm_2',
   *   condition: { optionIndices: [0], operator: 'any_of' },
   * })
   *
   * // Yes/No question: reveal when the respondent answers "yes".
   * const yesNo = await client.forms.addLogicRule('frm_123', 'elm_3', {
   *   revealedElementId: 'elm_4',
   *   condition: { selectionType: 'is', value: true },
   * })
   * ```
   */
  addLogicRule(
    id: string,
    elementId: string,
    body: FormAddLogicRuleBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Remove the logic rule that reveals `targetElementId` from a source
   * element. Returns the updated source element.
   *
   * @example
   * ```ts
   * const element = await client.forms.removeLogicRule('frm_123', 'elm_1', 'elm_2')
   * ```
   */
  removeLogicRule(
    id: string,
    elementId: string,
    targetElementId: string,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Iterate every form matching the filter, paging automatically.
   *
   * @example
   * ```ts
   * for await (const form of client.forms.listAutoPaginate({ type: 'standalone' })) {
   *   console.log(form.id, form.name)
   * }
   * ```
   */
  listAutoPaginate(
    params?: FormListParams,
    options?: RequestOptions,
  ): AsyncIterableIterator<FormSummary>
}

/**
 * Build a forms service bound to an {@link HttpClient}.
 *
 * @internal
 */
export function formsService(http: HttpClient): FormsService {
  return {
    async list(params: FormListParams = {}, options?: RequestOptions): Promise<ListFormsResponse> {
      return http.request<ListFormsResponse>({
        method: 'GET',
        path: '/forms',
        query: listParamsToQuery(params),
        options,
      })
    },

    async retrieve(id: string, options?: RequestOptions): Promise<Form> {
      requireString('retrieve', 'id', id)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'GET',
        path: `/forms/${encodeURIComponent(id)}`,
        options,
      })
      return response.data
    },

    async create(body: FormCreateBody, options?: RequestOptions): Promise<Form> {
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'POST',
        path: '/forms',
        body,
        options,
      })
      return response.data
    },

    async update(id: string, body: FormUpdateBody, options?: RequestOptions): Promise<Form> {
      requireString('update', 'id', id)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'PATCH',
        path: `/forms/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async duplicate(
      id: string,
      body: FormDuplicateBody = {},
      options?: RequestOptions,
    ): Promise<Form> {
      requireString('duplicate', 'id', id)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'POST',
        path: `/forms/${encodeURIComponent(id)}/duplicate`,
        body,
        options,
      })
      return response.data
    },

    async addElement(
      id: string,
      body: FormAddElementBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('addElement', 'id', id)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'POST',
        path: `/forms/${encodeURIComponent(id)}/elements`,
        body,
        options,
      })
      return response.data
    },

    async updateElement(
      id: string,
      elementId: string,
      body: FormUpdateElementBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('updateElement', 'id', id)
      requireString('updateElement', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'PATCH',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}`,
        body,
        options,
      })
      return response.data
    },

    async deleteElement(
      id: string,
      elementId: string,
      options?: RequestOptions,
    ): Promise<DeletedElement> {
      requireString('deleteElement', 'id', id)
      requireString('deleteElement', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<DeletedElement>>({
        method: 'DELETE',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}`,
        options,
      })
      return response.data
    },

    async moveElement(
      id: string,
      elementId: string,
      body: FormMoveElementBody,
      options?: RequestOptions,
    ): Promise<Form> {
      requireString('moveElement', 'id', id)
      requireString('moveElement', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'PUT',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}/position`,
        body,
        options,
      })
      return response.data
    },

    async enableOtherOption(
      id: string,
      elementId: string,
      body: FormEnableOtherOptionBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('enableOtherOption', 'id', id)
      requireString('enableOtherOption', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'PUT',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}/other-option`,
        body,
        options,
      })
      return response.data
    },

    async disableOtherOption(
      id: string,
      elementId: string,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('disableOtherOption', 'id', id)
      requireString('disableOtherOption', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'DELETE',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}/other-option`,
        options,
      })
      return response.data
    },

    async addLogicRule(
      id: string,
      elementId: string,
      body: FormAddLogicRuleBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('addLogicRule', 'id', id)
      requireString('addLogicRule', 'elementId', elementId)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'POST',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}/logic-rules`,
        body,
        options,
      })
      return response.data
    },

    async removeLogicRule(
      id: string,
      elementId: string,
      targetElementId: string,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireString('removeLogicRule', 'id', id)
      requireString('removeLogicRule', 'elementId', elementId)
      requireString('removeLogicRule', 'targetElementId', targetElementId)
      const response = await http.request<DetailEnvelope<FormElement>>({
        method: 'DELETE',
        path: `/forms/${encodeURIComponent(id)}/elements/${encodeURIComponent(elementId)}/logic-rules/${encodeURIComponent(targetElementId)}`,
        options,
      })
      return response.data
    },

    listAutoPaginate(
      params: FormListParams = {},
      options?: RequestOptions,
    ): AsyncIterableIterator<FormSummary> {
      const fetchPage = async (
        pageParams: FormListParams & { page: number },
      ): Promise<{ data: readonly FormSummary[]; hasMore: boolean }> => {
        const result = await http.request<ListFormsResponse>({
          method: 'GET',
          path: '/forms',
          query: listParamsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<FormSummary, FormListParams>(fetchPage, params)
    },
  }
}

function requireString(method: string, name: string, value: string): void {
  if (typeof value !== 'string' || value.length === 0) {
    throw new TypeError(`forms.${method}: \`${name}\` must be a non-empty string`)
  }
}

function listParamsToQuery(
  params: FormListParams,
): Record<string, string | number | boolean | undefined> {
  const query: Record<string, string | number | boolean | undefined> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined) continue
    if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
      query[key] = value
    }
  }
  return query
}
