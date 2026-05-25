import { readdir, readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import yaml from 'js-yaml'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { ThreeCommon } from '@/client'
import * as Errors from '@/errors'

import { setupMockServer, TEST_BASE_URL } from '../helpers/mock-server'

import { dispatchEvents } from './dispatch-events'
import { dispatchInvoices } from './dispatch-invoices'
import { dispatchSubscriptions } from './dispatch-subscriptions'

import type {
  ClientOverrides,
  ExpectedError,
  ExpectedRequest,
  MockResponse,
  Scenario,
  ScenarioCall,
} from './scenario'

const SCENARIOS_DIR = resolve(
  fileURLToPath(import.meta.url),
  '..',
  '..',
  '..',
  '..',
  'conformance',
  'scenarios',
)

async function loadScenarios(): Promise<readonly { file: string; scenario: Scenario }[]> {
  const files: string[] = []
  const walk = async (dir: string): Promise<void> => {
    const entries = await readdir(dir, { withFileTypes: true })
    for (const entry of entries) {
      const full = resolve(dir, entry.name)
      if (entry.isDirectory()) {
        await walk(full)
      } else if (entry.name.endsWith('.yaml')) {
        files.push(full)
      }
    }
  }
  await walk(SCENARIOS_DIR)
  files.sort()

  const scenarios = await Promise.all(
    files.map(async (full) => {
      const text = await readFile(full, 'utf-8')
      // Relative path (e.g. "events/list-happy.yaml") makes it obvious which
      // resource the scenario targets when scrolling test output.
      const file = full.slice(SCENARIOS_DIR.length + 1)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const scenario: any = yaml.load(text)
      return { file, scenario: scenario as Scenario }
    }),
  )
  return scenarios
}

const server = setupMockServer()

function buildClient(overrides: ClientOverrides = {}): ThreeCommon {
  return new ThreeCommon({
    apiKey: '3co_test',
    baseUrl: TEST_BASE_URL,
    apiVersion: overrides.apiVersion ?? '2026-04-29',
    telemetry: overrides.telemetry ?? true,
    maxRetries: overrides.maxRetries ?? 3,
    retryDelay: { initialMs: 1, maxMs: 4, jitter: false },
  })
}

// dispatchCall routes a scenario call to its per-resource dispatcher.
// Each resource lives in its own file (e.g. dispatch-events.ts,
// dispatch-invoices.ts); add a sibling case here when introducing a new
// resource.
function dispatchCall(
  client: ThreeCommon,
  call: ScenarioCall,
): Promise<unknown> | AsyncIterableIterator<unknown> {
  const resource = call.resource ?? 'events'
  switch (resource) {
    case 'events':
      return dispatchEvents(client, call)
    case 'invoices':
      return dispatchInvoices(client, call)
    case 'subscriptions':
      return dispatchSubscriptions(client, call)
  }
}

function assertRequest(
  observed: Request,
  observedBody: string,
  expected: ExpectedRequest,
  scenarioName: string,
): void {
  expect(observed.method, `[${scenarioName}] method`).toBe(expected.method)

  const url = new URL(observed.url)
  expect(url.pathname, `[${scenarioName}] path`).toBe(expected.path)

  if (expected.query !== undefined) {
    for (const [key, value] of Object.entries(expected.query)) {
      expect(url.searchParams.get(key), `[${scenarioName}] query.${key}`).toBe(value)
    }
  }

  if (expected.headers !== undefined) {
    for (const [key, value] of Object.entries(expected.headers)) {
      expect(observed.headers.get(key), `[${scenarioName}] header.${key}`).toBe(value)
    }
  }

  if (expected.headersAbsent !== undefined) {
    for (const key of expected.headersAbsent) {
      expect(observed.headers.has(key), `[${scenarioName}] header.${key} should be absent`).toBe(
        false,
      )
    }
  }

  if (expected.body !== undefined && expected.body !== null) {
    const parsed = observedBody.length > 0 ? (JSON.parse(observedBody) as unknown) : undefined
    expect(parsed, `[${scenarioName}] body`).toEqual(expected.body)
  }
}

function buildMockResponse(mock: MockResponse): Response {
  return HttpResponse.json(mock.body ?? null, {
    status: mock.status,
    headers: { 'Content-Type': 'application/json', ...(mock.headers ?? {}) },
  })
}

function assertError(err: unknown, expected: ExpectedError, scenarioName: string): void {
  const ExpectedClass = Errors[expected.type]
  expect(err, `[${scenarioName}] error type`).toBeInstanceOf(ExpectedClass)
  const e = err as Errors.ThreeCommonError & {
    retryAfterSeconds?: number
  }
  expect(e.code, `[${scenarioName}] error code`).toBe(expected.code)
  if (expected.httpStatus !== undefined) {
    expect(e.httpStatus, `[${scenarioName}] error httpStatus`).toBe(expected.httpStatus)
  }
  if (expected.requestId !== undefined) {
    expect(e.requestId, `[${scenarioName}] error requestId`).toBe(expected.requestId)
  }
  if (expected.retryAfterSeconds !== undefined) {
    expect(e.retryAfterSeconds, `[${scenarioName}] retryAfterSeconds`).toBe(
      expected.retryAfterSeconds,
    )
  }
  if (expected.details !== undefined) {
    expect(e.details, `[${scenarioName}] error details`).toEqual(expected.details)
  }
}

const allScenarios = await loadScenarios()

describe('cross-language conformance', () => {
  it('loads at least one scenario', () => {
    expect(allScenarios.length).toBeGreaterThan(0)
  })

  it.each(allScenarios)('$scenario.name', async ({ scenario }) => {
    const client = buildClient(scenario.client)
    const observedRequests: { req: Request; body: string }[] = []

    if (scenario.exchanges !== undefined) {
      let exchangeIndex = 0
      const exchanges = scenario.exchanges
      server.use(
        http.all(`${TEST_BASE_URL}${exchanges[0]?.request.path ?? ''}`, async ({ request }) => {
          const body = await request.clone().text()
          observedRequests.push({ req: request, body })
          const exchange = exchanges[exchangeIndex]
          exchangeIndex += 1
          if (exchange === undefined) {
            throw new Error(`unexpected extra request to ${request.url}`)
          }
          return buildMockResponse(exchange.response)
        }),
      )
    } else if (scenario.expectedRequest !== undefined && scenario.mockResponse !== undefined) {
      const expected = scenario.expectedRequest
      const mock = scenario.mockResponse
      server.use(
        http.all(`${TEST_BASE_URL}${expected.path}`, async ({ request }) => {
          const body = await request.clone().text()
          observedRequests.push({ req: request, body })
          return buildMockResponse(mock)
        }),
      )
    } else {
      throw new Error(
        `scenario "${scenario.name}" has no expectedRequest/mockResponse or exchanges`,
      )
    }

    let result: unknown
    let thrown: unknown
    try {
      const dispatched = dispatchCall(client, scenario.call)
      if (scenario.call.method === 'listAutoPaginate') {
        const collected: unknown[] = []
        for await (const item of dispatched as AsyncIterableIterator<unknown>) {
          collected.push(item)
        }
        result = collected
      } else {
        result = await dispatched
      }
    } catch (err) {
      thrown = err
    }

    if (scenario.expectedError !== undefined) {
      expect(thrown, `[${scenario.name}] expected an error`).toBeDefined()
      assertError(thrown, scenario.expectedError, scenario.name)
    } else if (thrown !== undefined) {
      throw thrown instanceof Error
        ? thrown
        : new Error(typeof thrown === 'string' ? thrown : 'unknown thrown value')
    }

    if (scenario.expectedRequest !== undefined && observedRequests[0] !== undefined) {
      assertRequest(
        observedRequests[0].req,
        observedRequests[0].body,
        scenario.expectedRequest,
        scenario.name,
      )
    }

    if (scenario.exchanges !== undefined) {
      scenario.exchanges.forEach((exchange, idx) => {
        const observed = observedRequests[idx]
        if (observed === undefined) {
          throw new Error(`[${scenario.name}] missing observed request at index ${String(idx)}`)
        }
        assertRequest(observed.req, observed.body, exchange.request, scenario.name)
      })
    }

    if (scenario.expectedResult !== undefined) {
      expect(result, `[${scenario.name}] result`).toMatchObject(scenario.expectedResult as object)
    }

    if (scenario.expectedAutoPaginated !== undefined) {
      const items = result as readonly unknown[]
      expect(items, `[${scenario.name}] auto-paginated count`).toHaveLength(
        scenario.expectedAutoPaginated.length,
      )
      scenario.expectedAutoPaginated.forEach((expectedItem, idx) => {
        expect(items[idx], `[${scenario.name}] auto-paginated[${String(idx)}]`).toMatchObject(
          expectedItem,
        )
      })
    }

    if (scenario.expectedCallCount !== undefined) {
      expect(observedRequests.length, `[${scenario.name}] call count`).toBe(
        scenario.expectedCallCount,
      )
    }
  })
})
