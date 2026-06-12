/**
 * Public types for the forms resource. Hand-curated friendly aliases over the
 * auto-generated OpenAPI types.
 *
 * @public
 */

import type { paths } from '@/generated/types'

/**
 * Lifecycle status of a form.
 *
 * - `draft`: not reachable at its public URL and hidden from most tables
 * - `active`: reachable at its URL and visible everywhere
 * - `archived`: effectively deleted, but restorable from the dashboard
 */
export type FormStatus = 'draft' | 'active' | 'archived'

/**
 * The kind of form.
 *
 * - `standalone`: a regular form not tied to event checkout
 * - `order`: drives an event checkout flow (buyer + ticket-holder sections)
 */
export type FormType = 'standalone' | 'order'

/**
 * A form in the compact projection returned by `list` and `listAutoPaginate`.
 * Contains only summary fields (id, name, element count, type, status).
 */
export type FormSummary = NonNullable<
  paths['/v1/forms/']['get']['responses'][200]['content']['application/json']['data']
>[number]

/**
 * The full form document returned by `create`, `retrieve`, `update`,
 * `duplicate`, and `moveElement`. A union of the standalone and order-form
 * shapes.
 */
export type Form = NonNullable<
  paths['/v1/forms/{formId}']['get']['responses'][200]['content']['application/json']['data']
>

/**
 * A single form element (question or static element) returned by
 * `addElement`, `updateElement`, `enableOtherOption`, `disableOtherOption`,
 * `addLogicRule`, and `removeLogicRule`. A union over every element type
 * (`Text`, `Select One`, `Yes/No`, image, etc.).
 */
export type FormElement = NonNullable<
  paths['/v1/forms/{formId}/elements']['post']['responses'][200]['content']['application/json']['data']
>

/** Result of `DELETE /v1/forms/{formId}/elements/{elementId}`. */
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
  /** If set, only forms of the given type are returned. Otherwise all forms. */
  readonly type?: FormType
}

/** Body accepted by `POST /v1/forms`. */
export type FormCreateBody =
  paths['/v1/forms/']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/forms/{formId}`. */
export type FormUpdateBody =
  paths['/v1/forms/{formId}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `POST /v1/forms/{formId}/duplicate`. */
export type FormDuplicateBody =
  paths['/v1/forms/{formId}/duplicate']['post']['requestBody']['content']['application/json']

/**
 * Body accepted by `POST /v1/forms/{formId}/elements`. A union over every
 * element type; the `type` discriminant selects the variant.
 */
export type FormAddElementBody =
  paths['/v1/forms/{formId}/elements']['post']['requestBody']['content']['application/json']

/** Body accepted by `PATCH /v1/forms/{formId}/elements/{elementId}`. */
export type FormUpdateElementBody =
  paths['/v1/forms/{formId}/elements/{elementId}']['patch']['requestBody']['content']['application/json']

/** Body accepted by `PUT /v1/forms/{formId}/elements/{elementId}/position`. */
export type FormMoveElementBody =
  paths['/v1/forms/{formId}/elements/{elementId}/position']['put']['requestBody']['content']['application/json']

/**
 * Body accepted by `PUT /v1/forms/{formId}/elements/{elementId}/other-option`
 * when enabling the free-text "Other" choice on a selection question.
 */
export type FormEnableOtherOptionBody =
  paths['/v1/forms/{formId}/elements/{elementId}/other-option']['put']['requestBody']['content']['application/json']

/**
 * Body accepted by `POST /v1/forms/{formId}/elements/{elementId}/logic-rules`.
 */
export type FormAddLogicRuleBody =
  paths['/v1/forms/{formId}/elements/{elementId}/logic-rules']['post']['requestBody']['content']['application/json']
