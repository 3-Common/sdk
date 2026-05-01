import {
  ThreeCommonAuthError,
  ThreeCommonConflictError,
  ThreeCommonNotFoundError,
  ThreeCommonPermissionError,
  ThreeCommonRateLimitError,
  ThreeCommonServerError,
  ThreeCommonValidationError,
} from './classes'

import type { ErrorResponseBody, ThreeCommonErrorInit, ThreeCommonError } from './base'

const STATUS_TO_SUBCLASS: Readonly<
  Record<number, new (init: ThreeCommonErrorInit) => ThreeCommonError>
> = {
  400: ThreeCommonValidationError,
  401: ThreeCommonAuthError,
  403: ThreeCommonPermissionError,
  404: ThreeCommonNotFoundError,
  409: ThreeCommonConflictError,
  422: ThreeCommonValidationError,
}

/**
 * Maps an HTTP response onto the appropriate {@link ThreeCommonError} subclass.
 *
 * @internal
 */
export function errorFromResponse(args: {
  readonly status: number
  readonly body: unknown
  readonly rawResponse: string | undefined
  readonly requestId: string | undefined
  readonly retryAfterSeconds: number | undefined
}): ThreeCommonError {
  const parsed = isErrorResponseBody(args.body) ? args.body.error : null
  const code = parsed?.code ?? defaultCodeForStatus(args.status)
  const message = parsed?.message ?? defaultMessageForStatus(args.status)
  const details = parsed?.details

  const init: ThreeCommonErrorInit = {
    code,
    message,
    httpStatus: args.status,
    requestId: args.requestId,
    details,
    rawResponse: args.rawResponse,
  }

  if (args.status === 429) {
    return new ThreeCommonRateLimitError({ ...init, retryAfterSeconds: args.retryAfterSeconds })
  }

  const Mapped = STATUS_TO_SUBCLASS[args.status]
  if (Mapped !== undefined) {
    return new Mapped(init)
  }

  if (args.status >= 500) {
    return new ThreeCommonServerError(init)
  }

  // Unknown 4xx — treat as validation by default; callers can inspect `httpStatus`.
  return new ThreeCommonValidationError(init)
}

function isErrorResponseBody(value: unknown): value is ErrorResponseBody {
  if (typeof value !== 'object' || value === null) return false
  const candidate = (value as { error?: unknown }).error
  if (typeof candidate !== 'object' || candidate === null) return false
  const errorObj = candidate as { code?: unknown; message?: unknown }
  return typeof errorObj.code === 'string' && typeof errorObj.message === 'string'
}

function defaultCodeForStatus(status: number): string {
  if (status === 401) return 'unauthorized'
  if (status === 403) return 'forbidden'
  if (status === 404) return 'not_found'
  if (status === 409) return 'conflict'
  if (status === 429) return 'rate_limit_exceeded'
  if (status >= 500) return 'internal_error'
  return 'request_failed'
}

function defaultMessageForStatus(status: number): string {
  return `Request failed with status ${String(status)}`
}
