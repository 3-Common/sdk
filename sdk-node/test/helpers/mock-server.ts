import { setupServer, type SetupServer } from 'msw/node'
import { afterAll, afterEach, beforeAll } from 'vitest'

import type { HttpHandler } from 'msw'

/**
 * Spin up an `msw` Node server scoped to a single Vitest file. Handlers can be
 * passed at setup time and overridden per-test via `server.use(...)`.
 */
export function setupMockServer(handlers: readonly HttpHandler[] = []): SetupServer {
  const server = setupServer(...handlers)
  beforeAll(() => {
    server.listen({ onUnhandledRequest: 'error' })
  })
  afterEach(() => {
    server.resetHandlers()
  })
  afterAll(() => {
    server.close()
  })
  return server
}

/** Convenience base URL used by every test. */
export const TEST_BASE_URL = 'https://api.test.3common.com'
