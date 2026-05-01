import process from 'node:process'

import { API_VERSION } from '@/api-version'
import { ThreeCommonValidationError } from '@/errors'

import type { ClientConfig, Logger } from '@/types/public'

/**
 * Resolved, validated configuration. Every field is required.
 *
 * @internal
 */
export interface ResolvedConfig {
  readonly apiKey: string
  readonly baseUrl: string
  readonly apiVersion: string
  readonly timeoutMs: number
  readonly maxRetries: number
  readonly retryDelay: {
    readonly initialMs: number
    readonly maxMs: number
    readonly jitter: boolean
  }
  readonly fetch: typeof fetch | undefined
  readonly logger: Logger | undefined
  readonly telemetry: boolean
}

const DEFAULTS = {
  baseUrl: 'https://api.3common.com',
  apiVersion: API_VERSION,
  timeoutMs: 30_000,
  maxRetries: 3,
  retryDelay: { initialMs: 500, maxMs: 8000, jitter: true },
  telemetry: true,
} as const

const ENV_VAR = 'THREECOMMON_API_KEY'

/**
 * Resolve a partial {@link ClientConfig} into a fully-specified
 * {@link ResolvedConfig}, applying defaults and validating required fields.
 *
 * @internal
 */
export function resolveConfig(input: ClientConfig): ResolvedConfig {
  const apiKey = input.apiKey ?? process.env[ENV_VAR]
  if (apiKey === undefined || apiKey.length === 0) {
    throw new ThreeCommonValidationError({
      code: 'missing_api_key',
      message: `An API key is required. Pass \`apiKey\` on the ThreeCommon constructor, or set the ${ENV_VAR} environment variable.`,
      httpStatus: undefined,
      requestId: undefined,
      details: undefined,
      rawResponse: undefined,
    })
  }

  const baseUrl = (input.baseUrl ?? DEFAULTS.baseUrl).replace(/\/+$/u, '')
  if (!/^https?:\/\//u.test(baseUrl)) {
    throw new ThreeCommonValidationError({
      code: 'invalid_base_url',
      message: `baseUrl must start with http:// or https://; got "${baseUrl}".`,
      httpStatus: undefined,
      requestId: undefined,
      details: undefined,
      rawResponse: undefined,
    })
  }

  const timeoutMs = input.timeoutMs ?? DEFAULTS.timeoutMs
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) {
    throw new ThreeCommonValidationError({
      code: 'invalid_timeout',
      message: `timeoutMs must be a positive number; got ${String(timeoutMs)}.`,
      httpStatus: undefined,
      requestId: undefined,
      details: undefined,
      rawResponse: undefined,
    })
  }

  const maxRetries = input.maxRetries ?? DEFAULTS.maxRetries
  if (!Number.isInteger(maxRetries) || maxRetries < 0) {
    throw new ThreeCommonValidationError({
      code: 'invalid_max_retries',
      message: `maxRetries must be a non-negative integer; got ${String(maxRetries)}.`,
      httpStatus: undefined,
      requestId: undefined,
      details: undefined,
      rawResponse: undefined,
    })
  }

  const retryDelay = {
    initialMs: input.retryDelay?.initialMs ?? DEFAULTS.retryDelay.initialMs,
    maxMs: input.retryDelay?.maxMs ?? DEFAULTS.retryDelay.maxMs,
    jitter: input.retryDelay?.jitter ?? DEFAULTS.retryDelay.jitter,
  }

  return {
    apiKey,
    baseUrl,
    apiVersion: input.apiVersion ?? DEFAULTS.apiVersion,
    timeoutMs,
    maxRetries,
    retryDelay,
    fetch: input.fetch,
    logger: input.logger,
    telemetry: input.telemetry ?? DEFAULTS.telemetry,
  }
}
