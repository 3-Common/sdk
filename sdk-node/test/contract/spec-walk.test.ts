import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

// Read spec.json (committed alongside spec.yaml) — parses without a library.
const SPEC_PATH = resolve(
  fileURLToPath(import.meta.url),
  '..',
  '..',
  '..',
  '..',
  'openapi',
  'spec.json',
)

interface OpenApiDocument {
  readonly openapi?: string
  readonly paths?: Readonly<Record<string, Readonly<Record<string, unknown>>>>
}

async function loadSpec(): Promise<OpenApiDocument> {
  const text = await readFile(SPEC_PATH, 'utf-8')
  return JSON.parse(text) as OpenApiDocument
}

describe('OpenAPI contract', () => {
  it('parses the canonical spec', async () => {
    const doc = await loadSpec()
    expect(doc.openapi).toMatch(/^3\./u)
    expect(doc.paths).toBeDefined()
  })

  it('declares the endpoints the SDK targets', async () => {
    const doc = await loadSpec()
    const paths = doc.paths ?? {}
    const expectedPaths = ['/v1/events/', '/v1/events/{id}']
    for (const path of expectedPaths) {
      expect(paths, `spec is missing ${path}`).toHaveProperty(path)
    }
  })

  it('all event endpoints declare bearer auth', async () => {
    const doc = await loadSpec()
    const paths = doc.paths ?? {}
    for (const [pathKey, methods] of Object.entries(paths)) {
      if (!pathKey.startsWith('/v1/events')) continue
      for (const [method, op] of Object.entries(methods)) {
        if (!['get', 'post', 'patch', 'put', 'delete'].includes(method)) continue
        const security = (op as { security?: readonly Record<string, unknown>[] }).security
        expect(security, `${method} ${pathKey} should require auth`).toBeDefined()
        expect(security?.[0]).toHaveProperty('bearerAuth')
      }
    }
  })
})
