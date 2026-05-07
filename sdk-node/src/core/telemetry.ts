import { API_VERSION } from '@/api-version'
import { SDK_VERSION } from '@/version'

/**
 * Telemetry header builder + last-request tracker.
 *
 * The header name `Threecommon-Client-Telemetry`
 * Disable globally with `telemetry: false` on client config or at runtime via
 *`client.disableTelemetry()`.
 *
 * @internal
 */

/**
 * Snapshot of the previous request, attached to the next request as a header.
 *
 * @internal
 */
export interface LastRequestMetric {
  readonly method: string
  readonly path: string
  readonly status: number | undefined
  readonly durationMs: number
  readonly requestId: string | undefined
}

/**
 * Stateful telemetry tracker. One instance per {@link ThreeCommon} client.
 *
 * @internal
 */
export class Telemetry {
  private enabledFlag: boolean
  private last: LastRequestMetric | undefined

  public constructor(enabled: boolean) {
    this.enabledFlag = enabled
    this.last = undefined
  }

  public isEnabled(): boolean {
    return this.enabledFlag
  }

  public disable(): void {
    this.enabledFlag = false
    this.last = undefined
  }

  /** Record a completed request. No-op when disabled. */
  public record(metric: LastRequestMetric): void {
    if (!this.enabledFlag) return
    this.last = metric
  }

  /**
   * Build the value for the next request's `Threecommon-Client-Telemetry`
   * header, or return `undefined` if telemetry is disabled.
   */
  public buildHeaderValue(): string | undefined {
    if (!this.enabledFlag) return undefined

    const payload = {
      lang: 'node',
      sdk: SDK_VERSION,
      api: API_VERSION,
      last:
        this.last === undefined
          ? undefined
          : {
              m: this.last.method,
              p: this.last.path,
              s: this.last.status,
              d: this.last.durationMs,
            },
    }

    return JSON.stringify(payload)
  }
}
