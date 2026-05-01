/**
 * Header builder. Takes already-resolved values and returns a
 * `Headers` object ready to attach to a `fetch` call.
 *
 * @internal
 */
export function buildHeaders(args: {
  readonly apiKey: string
  readonly apiVersion: string
  readonly sdkVersion: string
  readonly userAgentSuffix: string
  readonly telemetryHeader: string | undefined
  readonly idempotencyKey: string | undefined
}): Headers {
  const headers = new Headers()
  headers.set('Authorization', `Bearer ${args.apiKey}`)
  headers.set('Threecommon-Version', args.apiVersion)
  headers.set('User-Agent', `ThreeCommonNode/${args.sdkVersion} (${args.userAgentSuffix})`)
  headers.set('Accept', 'application/json')
  headers.set('Content-Type', 'application/json')

  if (args.telemetryHeader !== undefined) {
    headers.set('Threecommon-Client-Telemetry', args.telemetryHeader)
  }

  if (args.idempotencyKey !== undefined) {
    headers.set('Idempotency-Key', args.idempotencyKey)
  }

  return headers
}
