"""Iter / AsyncIter tests."""

from __future__ import annotations

import pytest

from threecommon.pagination import AsyncIter, Iter


def test_iter_walks_multiple_pages() -> None:
    pages: list[list[int]] = [[1, 2], [3]]
    calls = 0

    def fetch(page: int) -> tuple[list[int], bool]:
        nonlocal calls
        calls += 1
        if page >= len(pages):
            return [], False
        return pages[page], page < len(pages) - 1

    out = list(Iter(fetch_page=fetch))
    assert out == [1, 2, 3]
    assert calls == 2


def test_iter_stops_on_empty_first_page() -> None:
    out: list[int] = list(Iter[int](fetch_page=lambda _p: ([], False)))
    assert out == []


def test_iter_propagates_error() -> None:
    def fetch(page: int) -> tuple[list[int], bool]:
        if page == 0:
            return [1], True
        msg = "network"
        raise RuntimeError(msg)

    iter_: Iter[int] = Iter(fetch_page=fetch)
    assert next(iter_) == 1
    with pytest.raises(RuntimeError, match="network"):
        next(iter_)


def test_iter_start_page() -> None:
    seen: list[int] = []

    def fetch(page: int) -> tuple[list[int], bool]:
        seen.append(page)
        return [page * 10], False

    out = list(Iter(fetch_page=fetch, start_page=3))
    assert out == [30]
    assert seen == [3]


@pytest.mark.asyncio
async def test_async_iter_walks_pages() -> None:
    pages: list[list[str]] = [["a", "b"], ["c"]]

    async def fetch(page: int) -> tuple[list[str], bool]:
        if page >= len(pages):
            return [], False
        return pages[page], page < len(pages) - 1

    got: list[str] = []
    async for v in AsyncIter(fetch_page=fetch):
        got.append(v)
    assert got == ["a", "b", "c"]


@pytest.mark.asyncio
async def test_async_iter_propagates_error() -> None:
    async def fetch(page: int) -> tuple[list[int], bool]:
        if page == 0:
            return [1, 2], True
        msg = "network"
        raise RuntimeError(msg)

    it = AsyncIter(fetch_page=fetch)
    assert await it.__anext__() == 1
    assert await it.__anext__() == 2
    with pytest.raises(RuntimeError, match="network"):
        await it.__anext__()


@pytest.mark.asyncio
async def test_async_iter_empty() -> None:
    async def fetch(_p: int) -> tuple[list[int], bool]:
        return [], False

    out = [v async for v in AsyncIter(fetch_page=fetch)]
    assert out == []
