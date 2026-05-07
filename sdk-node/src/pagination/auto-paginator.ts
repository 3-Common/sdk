/**
 * Auto-paginator factory. Wraps a "fetch one page" function in an
 * `AsyncIterableIterator<T>` that walks every page until the API reports
 * `hasMore: false`.
 *
 * @internal
 */

export interface Page<T> {
  readonly data: readonly T[]
  readonly hasMore: boolean
}

/**
 * Build an async iterable that walks pages from `fetchPage`.
 *
 * @internal
 */
export function createAutoPaginator<T, P extends { page?: number }>(
  fetchPage: (params: P & { page: number }) => Promise<Page<T>>,
  initialParams: P,
): AsyncIterableIterator<T> {
  let pageNumber = initialParams.page ?? 0
  let buffer: readonly T[] = []
  let bufferIndex = 0
  let hasMore = true

  const iterator: AsyncIterableIterator<T> = {
    [Symbol.asyncIterator](): AsyncIterableIterator<T> {
      return iterator
    },
    async next(): Promise<IteratorResult<T>> {
      if (bufferIndex < buffer.length) {
        const value = buffer[bufferIndex]
        bufferIndex += 1
        return { value: value as T, done: false }
      }

      if (!hasMore) {
        return { value: undefined, done: true }
      }

      const params = { ...initialParams, page: pageNumber } as P & { page: number }
      const page = await fetchPage(params)
      buffer = page.data
      bufferIndex = 0
      hasMore = page.hasMore
      pageNumber += 1

      if (buffer.length === 0) {
        return { value: undefined, done: true }
      }

      const first = buffer[bufferIndex]
      bufferIndex += 1
      return { value: first as T, done: false }
    },
  }

  return iterator
}
