/**
 * Public types for the forms resource. Hand-curated friendly aliases over
 * the auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * The kind of form.
 *
 * - `standalone`: a self-contained form accessed at its own URL
 * - `order`: an order form attached to an event's checkout flow
 *
 * Converting between the two is a destructive operation server-side.
 */
export type FormType = 'standalone' | 'order'

/**
 * Lifecycle status of a form.
 *
 * - `draft`: not reachable at its URL and hidden from most dashboard tables
 * - `active`: reachable at its URL and visible everywhere
 * - `archived`: effectively deleted, but restorable from the forms dashboard
 */
export type FormStatus = 'draft' | 'active' | 'archived'

/**
 * Horizontal alignment of a form's submit button, applied when
 * `submitButtonWidth` is `auto`.
 */
export type SubmitButtonAlign = 'left' | 'center'

/**
 * A full form record, as returned by `retrieve`, `create`, `update`,
 * `duplicate`, and `moveElement`. The shape is a union of the standalone and
 * order-form projections; order forms additionally carry `attendeeRowsStart`.
 */
export type Form = NonNullable<
  paths['/v1/forms/{formId}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * A single form element (question, content block, or image). Returned by
 * `addElement`, `updateElement`, `enableOtherOption`, `disableOtherOption`,
 * and `addLogicRule`. The concrete fields depend on the element's `type`.
 */
export type FormElement = NonNullable<
  paths['/v1/forms/{formId}/elements']['post']['responses'][200]['content']['application/json']['data']
>

/**
 * The compact form projection returned by `list`. Carries just enough to
 * render a list row: `id`, `name`, `numElements`, `type`, and `status`.
 */
export type FormSummary = NonNullable<
  paths['/v1/forms/']['get']['responses'][200]['content']['application/json']['data']
>[number]

/** Result of `DELETE /v1/forms/{formId}/elements/{elementId}` -- the id of the removed element. */
export type DeletedElement = NonNullable<
  paths['/v1/forms/{formId}/elements/{elementId}']['delete']['responses'][200]['content']['application/json']['data']
>

/** Successful response shape from `GET /v1/forms`. */
export interface ListFormsResponse {
  readonly data: readonly FormSummary[]
  readonly hasMore: boolean
}

/** Query parameters accepted by `GET /v1/forms`. */
export interface FormListParams {
  /** Page number, 0-indexed. Default 0. */
  readonly page?: number
  /** Items per page. Default 50, max 250. */
  readonly pageSize?: number
  /** Restrict to a single form type; omit to include all. */
  readonly type?: FormType
}

/** Body accepted by `POST /v1/forms`. */
export type FormCreateBody =
  paths['/v1/forms/']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/forms/{formId}`. Only fields you provide are changed. */
export type FormUpdateBody =
  paths['/v1/forms/{formId}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/forms/{formId}/duplicate`. */
export type FormDuplicateBody =
  paths['/v1/forms/{formId}/duplicate']['post']['requestBody']['content']['application/json']

/**
 * Body accepted by `POST /v1/forms/{formId}/elements`. A discriminated union
 * over the element `type`; each variant accepts the fields valid for that
 * question kind.
 */
export type AddElementBody =
  paths['/v1/forms/{formId}/elements']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/forms/{formId}/elements/{elementId}`. */
export type UpdateElementBody =
  paths['/v1/forms/{formId}/elements/{elementId}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `PUT /v1/forms/{formId}/elements/{elementId}/position`. */
export type MoveElementBody =
  paths['/v1/forms/{formId}/elements/{elementId}/position']['put']['requestBody']['content']['application/json']

/** Body accepted by `PUT /v1/forms/{formId}/elements/{elementId}/other-option`. */
export type EnableOtherOptionBody =
  paths['/v1/forms/{formId}/elements/{elementId}/other-option']['put']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/forms/{formId}/elements/{elementId}/logic-rules`. */
export type AddLogicRuleBody =
  paths['/v1/forms/{formId}/elements/{elementId}/logic-rules']['post']['requestBody']['content']['application/json']
