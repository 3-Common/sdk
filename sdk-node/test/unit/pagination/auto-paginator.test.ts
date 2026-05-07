import { describe, expect, it } from 'vitest'

import { createAutoPaginator, type Page } from '@/pagination/auto-paginator'

describe('createAutoPaginator', () => {
  it('walks every page until hasMore is false', async () => {
    const pages: readonly Page<number>[] = [
      { data: [1, 2], hasMore: true },
      { data: [3, 4], hasMore: true },
      { data: [5], hasMore: false },
    ]

    const fetchPage = async (params: { page: number }): Promise<Page<number>> => {
      await Promise.resolve()
      const idx = params.page
      return pages[idx] ?? { data: [], hasMore: false }
    }

    const seen: number[] = []
    for await (const value of createAutoPaginator<number, { page?: number }>(fetchPage, {})) {
      seen.push(value)
    }
    expect(seen).toEqual([1, 2, 3, 4, 5])
  })

  it('returns no items when first page is empty', async () => {
    const fetchPage = async (): Promise<Page<number>> => {
      await Promise.resolve()
      return { data: [], hasMore: false }
    }
    const seen: number[] = []
    for await (const value of createAutoPaginator<number, { page?: number }>(fetchPage, {})) {
      seen.push(value)
    }
    expect(seen).toEqual([])
  })

  it('respects starting page when provided', async () => {
    const calls: number[] = []
    const fetchPage = async (params: { page: number }): Promise<Page<number>> => {
      await Promise.resolve()
      calls.push(params.page)
      return { data: [params.page * 10], hasMore: params.page < 3 }
    }
    const seen: number[] = []
    for await (const value of createAutoPaginator<number, { page?: number }>(fetchPage, {
      page: 2,
    })) {
      seen.push(value)
    }
    expect(calls).toEqual([2, 3])
    expect(seen).toEqual([20, 30])
  })

  it('forwards initial params verbatim alongside the incrementing page', async () => {
    let observed: { page: number; status?: string } = { page: -1 }
    const fetchPage = async (params: { page: number; status?: string }): Promise<Page<number>> => {
      observed = params
      await Promise.resolve()
      return { data: [1], hasMore: false }
    }
    const iter = createAutoPaginator<number, { page?: number; status?: string }>(fetchPage, {
      status: 'open',
    })
    await iter.next()
    expect(observed).toEqual({ page: 0, status: 'open' })
  })
})
