/**
 * Error model
 *
 * Every error thrown by the SDK is a {@link ThreeCommonError}. Subclasses are
 * keyed off the API's `error.code` field plus the HTTP status.
 */

/**
 * Shape of the JSON body the API returns on errors.
 *
 * @public
 */
export interface ErrorResponseBody {
  readonly error: {
    readonly code: string
    readonly message: string
    readonly details?: Readonly<Record<string, unknown>>
  }
}

/**
 * Constructor input shared by every {@link ThreeCommonError} subclass.
 *
 * @internal
 */
export interface ThreeCommonErrorInit {
  readonly code: string
  readonly message: string
  readonly httpStatus: number | undefined
  readonly requestId: string | undefined
  readonly details: Readonly<Record<string, unknown>> | undefined
  readonly rawResponse: string | undefined
  readonly cause?: unknown
}

/**
 * Base class for every error thrown by this SDK.
 *
 * @public
 */
export abstract class ThreeCommonError extends Error {
  /** Stable string identifier matching the API's `error.code`. */
  public readonly code: string
  /** HTTP status code, when the error originated from a response. */
  public readonly httpStatus: number | undefined
  /** Request ID from the `X-Request-ID` response header, when present. */
  public readonly requestId: string | undefined
  /** Optional additional context from the API's `error.details` field. */
  public readonly details: Readonly<Record<string, unknown>> | undefined
  /** The raw response body, when retained for debugging. */
  public readonly rawResponse: string | undefined

  protected constructor(init: ThreeCommonErrorInit) {
    super(init.message, init.cause === undefined ? undefined : { cause: init.cause })
    this.name = new.target.name
    this.code = init.code
    this.httpStatus = init.httpStatus
    this.requestId = init.requestId
    this.details = init.details
    this.rawResponse = init.rawResponse
  }

  /**
   * Single-line representation suitable for logs. Includes the request ID so
   * customer support can grep their server logs to correlate.
   */
  public override toString(): string {
    const id = this.requestId === undefined ? '' : ` (request_id=${this.requestId})`
    return `[${this.code}] ${this.message}${id}`
  }
}
