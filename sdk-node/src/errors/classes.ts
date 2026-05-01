import { ThreeCommonError, type ThreeCommonErrorInit } from './base'

/** 401 Unauthorized — invalid, missing, or expired API key. */
export class ThreeCommonAuthError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/** 403 Forbidden — the API key lacks the required scope for this endpoint. */
export class ThreeCommonPermissionError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/** 404 Not Found — the requested resource does not exist. */
export class ThreeCommonNotFoundError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/** 400 / 422 — request validation failed. */
export class ThreeCommonValidationError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/** 409 Conflict — request conflicts with current resource state. */
export class ThreeCommonConflictError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/**
 * 429 Too Many Requests. Carries `retryAfterSeconds` parsed from the
 * `Retry-After` header so callers can implement their own backoff.
 */
export class ThreeCommonRateLimitError extends ThreeCommonError {
  public readonly retryAfterSeconds: number | undefined

  /** @internal */
  public constructor(
    init: ThreeCommonErrorInit & { readonly retryAfterSeconds: number | undefined },
  ) {
    super(init)
    this.retryAfterSeconds = init.retryAfterSeconds
  }
}

/** 5xx — the API returned an unexpected server-side failure. */
export class ThreeCommonServerError extends ThreeCommonError {
  /** @internal */
  public constructor(init: ThreeCommonErrorInit) {
    super(init)
  }
}

/**
 * Network-level failure — DNS resolution, TCP reset, TLS error, request abort
 * after timeout, etc. Carries the underlying error as `cause`.
 */
export class ThreeCommonConnectionError extends ThreeCommonError {
  /** @internal */
  public constructor(
    init: Omit<ThreeCommonErrorInit, 'httpStatus' | 'rawResponse'> & { readonly cause: unknown },
  ) {
    super({ ...init, httpStatus: undefined, rawResponse: undefined })
  }
}
