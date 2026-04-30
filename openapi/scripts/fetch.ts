/**
 * Fetches the canonical OpenAPI spec from the 3Common API and writes
 * the bundled JSON + YAML snapshots into ../spec.json / ../spec.yaml.
 *
 * Run from this directory:
 *   npx tsx scripts/fetch.ts
 */

import { writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import yaml from 'js-yaml'

const DEFAULT_SOURCE = 'https://api.3common.com/docs/json'
const SOURCE_URL = process.env.OPENAPI_SOURCE_URL ?? DEFAULT_SOURCE

const OUT_DIR = resolve(import.meta.dirname, '..')
const JSON_PATH = resolve(OUT_DIR, 'spec.json')
const YAML_PATH = resolve(OUT_DIR, 'spec.yaml')

async function main(): Promise<void> {
  process.stdout.write(`Fetching ${SOURCE_URL}\n`)
  const response = await fetch(SOURCE_URL, {
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    throw new Error(`Source returned ${response.status} ${response.statusText}`)
  }
  const spec = (await response.json()) as Record<string, unknown>

  if (typeof spec.openapi !== 'string') {
    throw new Error("Response is not an OpenAPI document (missing 'openapi' field)")
  }

  const json = `${JSON.stringify(spec, null, 2)}\n`
  const yml = yaml.dump(spec, { lineWidth: 120, noRefs: true, sortKeys: false })

  await writeFile(JSON_PATH, json, 'utf-8')
  await writeFile(YAML_PATH, yml, 'utf-8')

  process.stdout.write(`Wrote ${JSON_PATH}\n`)
  process.stdout.write(`Wrote ${YAML_PATH}\n`)
}

main().catch((err: unknown) => {
  process.stderr.write(`fetch.ts failed: ${err instanceof Error ? err.message : String(err)}\n`)
  process.exit(1)
})
