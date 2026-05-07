import { describe, expect, it } from 'vitest'

import { buildUrl } from '@/core/url'

describe('buildUrl', () => {
  const base = 'https://api.test.example.com'

  it('joins baseUrl + apiPath + path', () => {
    expect(buildUrl({ baseUrl: base, apiPath: '/v1', path: '/events', query: undefined })).toBe(
      `${base}/v1/events`,
    )
  })

  it('strips trailing slashes from baseUrl', () => {
    expect(
      buildUrl({ baseUrl: `${base}///`, apiPath: '/v1', path: '/events', query: undefined }),
    ).toBe(`${base}/v1/events`)
  })

  it('prefixes a leading slash to path when missing', () => {
    expect(buildUrl({ baseUrl: base, apiPath: '/v1', path: 'events', query: undefined })).toBe(
      `${base}/v1/events`,
    )
  })

  it('appends a query string when query is non-empty', () => {
    const url = buildUrl({
      baseUrl: base,
      apiPath: '/v1',
      path: '/events',
      query: { status: 'open', pageSize: 10, includeArchived: true, missing: undefined },
    })
    expect(url).toContain('status=open')
    expect(url).toContain('pageSize=10')
    expect(url).toContain('includeArchived=true')
    expect(url).not.toContain('missing')
  })

  it('omits the query string when all values are undefined', () => {
    const url = buildUrl({
      baseUrl: base,
      apiPath: '/v1',
      path: '/events',
      query: { missing: undefined },
    })
    expect(url).toBe(`${base}/v1/events`)
  })
})
