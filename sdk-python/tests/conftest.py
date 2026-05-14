"""Shared pytest fixtures."""

from __future__ import annotations

from collections.abc import Iterator

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon


@pytest.fixture
def sync_client() -> Iterator[ThreeCommon]:
    """A sync client wired to a non-routable base — pytest-httpx intercepts the wire."""
    client = ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)
    yield client
    client.close()


@pytest.fixture
async def async_client() -> AsyncThreeCommon:
    """An async client wired to a non-routable base — pytest-httpx intercepts."""
    client = AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)
    try:
        return client
    finally:
        # The fixture isn't a context manager because we want the test to
        # close manually if needed; pytest-httpx tears down the transport.
        pass
