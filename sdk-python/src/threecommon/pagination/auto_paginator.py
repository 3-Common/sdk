"""auto-paginating iterators returned by every list endpoint.

Two flavors:

* [Iter][threecommon.pagination.Iter] — sync iterator. Use ``for ev in iter:``.
* [AsyncIter][threecommon.pagination.AsyncIter] — async iterator.
  Use ``async for ev in iter:``.

Both walk pages lazily — one HTTP call per page, only when the user
drains the previous page's buffer.
"""

from __future__ import annotations

from collections.abc import AsyncIterator, Awaitable, Callable, Iterator
from typing import Generic, TypeVar

T = TypeVar("T")

#: Synchronous page-fetch callback. Returns ``(page_items, has_more)``.
SyncFetchPage = Callable[[int], tuple[list[T], bool]]

#: Asynchronous page-fetch callback. Returns ``(page_items, has_more)``.
AsyncFetchPage = Callable[[int], Awaitable[tuple[list[T], bool]]]


class Iter(Generic[T]):
    """Sync auto-paginating iterator. ``for ev in iter:`` drives the pages."""

    __slots__ = ("_buffer", "_fetch_page", "_has_more", "_index", "_page")

    def __init__(self, *, fetch_page: SyncFetchPage[T], start_page: int = 0) -> None:
        self._fetch_page = fetch_page
        self._page = start_page
        self._buffer: list[T] = []
        self._index = 0
        self._has_more = True

    def __iter__(self) -> Iterator[T]:
        return self

    def __next__(self) -> T:
        if self._index < len(self._buffer):
            value = self._buffer[self._index]
            self._index += 1
            return value
        if not self._has_more:
            raise StopIteration

        data, has_more = self._fetch_page(self._page)
        self._buffer = data
        self._index = 0
        self._has_more = has_more
        self._page += 1

        if not self._buffer:
            raise StopIteration

        value = self._buffer[self._index]
        self._index += 1
        return value


class AsyncIter(Generic[T]):
    """Async auto-paginating iterator. ``async for ev in iter:`` drives the pages."""

    __slots__ = ("_buffer", "_fetch_page", "_has_more", "_index", "_page")

    def __init__(self, *, fetch_page: AsyncFetchPage[T], start_page: int = 0) -> None:
        self._fetch_page = fetch_page
        self._page = start_page
        self._buffer: list[T] = []
        self._index = 0
        self._has_more = True

    def __aiter__(self) -> AsyncIterator[T]:
        return self

    async def __anext__(self) -> T:
        if self._index < len(self._buffer):
            value = self._buffer[self._index]
            self._index += 1
            return value
        if not self._has_more:
            raise StopAsyncIteration

        data, has_more = await self._fetch_page(self._page)
        self._buffer = data
        self._index = 0
        self._has_more = has_more
        self._page += 1

        if not self._buffer:
            raise StopAsyncIteration

        value = self._buffer[self._index]
        self._index += 1
        return value
