import { createAutoPaginator } from '@/pagination/auto-paginator'

import type {
  AddElementBody,
  AddLogicRuleBody,
  DeletedElement,
  EnableOtherOptionBody,
  Form,
  FormCreateBody,
  FormDuplicateBody,
  FormElement,
  FormListParams,
  FormSummary,
  FormUpdateBody,
  ListFormsResponse,
  MoveElementBody,
  UpdateElementBody,
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
   * Create a new form. Requires a `name` and a `type`.
   *
   * @example
   * ```ts
   * const form = await client.forms.create({ name: 'Customer survey', type: 'standalone' })
   * ```
   */
  create(body: FormCreateBody, options?: RequestOptions): Promise<Form>

  /**
   * Retrieve a single form, including its rows and elements, by id.
   *
   * @example
   * ```ts
   * const form = await client.forms.retrieve('frm_123')
   * ```
   */
  retrieve(id: string, options?: RequestOptions): Promise<Form>

  /**
   * Edit a form's top-level settings (name, status, submit-button options).
   * Only the fields you provide are changed.
   *
   * @example
   * ```ts
   * const updated = await client.forms.update('frm_123', { name: 'Renamed survey', status: 'active' })
   * ```
   */
  update(id: string, body: FormUpdateBody, options?: RequestOptions): Promise<Form>

  /**
   * Duplicate an existing form, optionally renaming the copy and setting its
   * initial status.
   *
   * @example
   * ```ts
   * const copy = await client.forms.duplicate('frm_123', { name: 'Customer survey (copy)' })
   * ```
   */
  duplicate(id: string, body: FormDuplicateBody, options?: RequestOptions): Promise<Form>

  /**
   * Add an element (question, content block, or image) to a form. The body is
   * a discriminated union over the element `type`.
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
  addElement(id: string, body: AddElementBody, options?: RequestOptions): Promise<FormElement>

  /**
   * Edit an existing element on a form. Only the fields you provide change.
   *
   * @example
   * ```ts
   * const element = await client.forms.updateElement('frm_123', 'elm_123', {
   *   prompt: 'What is your full name?',
   * })
   * ```
   */
  updateElement(
    id: string,
    elementId: string,
    body: UpdateElementBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Delete an element from a form. Returns `{ deletedElementId }` echoing the
   * removed element's id.
   *
   * @example
   * ```ts
   * await client.forms.deleteElement('frm_123', 'elm_123')
   * ```
   */
  deleteElement(id: string, elementId: string, options?: RequestOptions): Promise<DeletedElement>

  /**
   * Move an element to a new position within the form. Returns the updated
   * form so the caller can observe the new ordering.
   *
   * @example
   * ```ts
   * const form = await client.forms.moveElement('frm_123', 'elm_123', { position: 2 })
   * ```
   */
  moveElement(
    id: string,
    elementId: string,
    body: MoveElementBody,
    options?: RequestOptions,
  ): Promise<Form>

  /**
   * Enable the "Other" free-text option on a selection element, setting its
   * prompt. Only valid on selection-style questions.
   *
   * @example
   * ```ts
   * const element = await client.forms.enableOtherOption('frm_123', 'elm_select', {
   *   otherPrompt: 'Other (please specify)',
   * })
   * ```
   */
  enableOtherOption(
    id: string,
    elementId: string,
    body: EnableOtherOptionBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Disable the "Other" free-text option on a selection element.
   *
   * @example
   * ```ts
   * const element = await client.forms.disableOtherOption('frm_123', 'elm_select')
   * ```
   */
  disableOtherOption(id: string, elementId: string, options?: RequestOptions): Promise<FormElement>

  /**
   * Add a conditional-logic rule to a selection or Yes/No element: when the
   * configured condition is met, the `revealedElementId` element is shown.
   * Returns the source element with its updated logic groups.
   *
   * @example
   * ```ts
   * const element = await client.forms.addLogicRule('frm_123', 'elm_select', {
   *   revealedElementId: 'elm_followup',
   *   condition: { optionIndices: [0], operator: 'any_of' },
   * })
   * ```
   */
  addLogicRule(
    id: string,
    elementId: string,
    body: AddLogicRuleBody,
    options?: RequestOptions,
  ): Promise<FormElement>

  /**
   * Remove the logic rule on `elementId` that reveals `targetElementId`.
   * Returns the source element with its remaining logic groups.
   *
   * @example
   * ```ts
   * const element = await client.forms.removeLogicRule('frm_123', 'elm_select', 'elm_followup')
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
   *   console.log(form.name)
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
        query: paramsToQuery(params),
        options,
      })
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

    async retrieve(id: string, options?: RequestOptions): Promise<Form> {
      requireArg('retrieve', 'id', id)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'GET',
        path: `/forms/${encodeURIComponent(id)}`,
        options,
      })
      return response.data
    },

    async update(id: string, body: FormUpdateBody, options?: RequestOptions): Promise<Form> {
      requireArg('update', 'id', id)
      const response = await http.request<DetailEnvelope<Form>>({
        method: 'PATCH',
        path: `/forms/${encodeURIComponent(id)}`,
        body,
        options,
      })
      return response.data
    },

    async duplicate(id: string, body: FormDuplicateBody, options?: RequestOptions): Promise<Form> {
      requireArg('duplicate', 'id', id)
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
      body: AddElementBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireArg('addElement', 'id', id)
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
      body: UpdateElementBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireArg('updateElement', 'id', id)
      requireArg('updateElement', 'elementId', elementId)
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
      requireArg('deleteElement', 'id', id)
      requireArg('deleteElement', 'elementId', elementId)
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
      body: MoveElementBody,
      options?: RequestOptions,
    ): Promise<Form> {
      requireArg('moveElement', 'id', id)
      requireArg('moveElement', 'elementId', elementId)
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
      body: EnableOtherOptionBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireArg('enableOtherOption', 'id', id)
      requireArg('enableOtherOption', 'elementId', elementId)
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
      requireArg('disableOtherOption', 'id', id)
      requireArg('disableOtherOption', 'elementId', elementId)
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
      body: AddLogicRuleBody,
      options?: RequestOptions,
    ): Promise<FormElement> {
      requireArg('addLogicRule', 'id', id)
      requireArg('addLogicRule', 'elementId', elementId)
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
      requireArg('removeLogicRule', 'id', id)
      requireArg('removeLogicRule', 'elementId', elementId)
      requireArg('removeLogicRule', 'targetElementId', targetElementId)
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
          query: paramsToQuery(pageParams),
          options,
        })
        return { data: result.data, hasMore: result.hasMore }
      }
      return createAutoPaginator<FormSummary, FormListParams>(fetchPage, params)
    },
  }
}

function requireArg(method: string, name: string, value: string): void {
  if (typeof value !== 'string' || value.length === 0) {
    throw new TypeError(`forms.${method}: \`${name}\` must be a non-empty string`)
  }
}

function paramsToQuery(
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
