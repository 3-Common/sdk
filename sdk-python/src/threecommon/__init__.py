"""Official Python client for the 3Common Public API.

Top-level entry points:

* [ThreeCommon][threecommon.ThreeCommon] — synchronous client
* [AsyncThreeCommon][threecommon.AsyncThreeCommon] — async client (httpx-backed)

Quick start:

    from threecommon import ThreeCommon

    client = ThreeCommon(api_key="3co_...")
    result = client.events.list(status="open", page_size=50)

    for event in client.events.list_auto_paginate(status="open"):
        print(event.name)

Async equivalent:

    import asyncio
    from threecommon import AsyncThreeCommon

    async def main() -> None:
        async with AsyncThreeCommon(api_key="3co_...") as client:
            async for event in client.events.list_auto_paginate(status="open"):
                print(event.name)

    asyncio.run(main())
"""

from threecommon.api_version import API_PATH, API_VERSION
from threecommon.client import AsyncThreeCommon, ThreeCommon
from threecommon.config import (
    DEFAULT_BASE_URL,
    DEFAULT_MAX_RETRIES,
    DEFAULT_RETRY_DELAY,
    DEFAULT_TIMEOUT_SECONDS,
    ClientConfig,
    RetryDelay,
)
from threecommon.errors.base import APIError
from threecommon.errors.classes import (
    AuthError,
    ConflictError,
    ConnectionError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ValidationError,
)
from threecommon.version import VERSION, __version__

__all__ = (
    # Constants
    "API_PATH",
    "API_VERSION",
    "DEFAULT_BASE_URL",
    "DEFAULT_MAX_RETRIES",
    "DEFAULT_RETRY_DELAY",
    "DEFAULT_TIMEOUT_SECONDS",
    "VERSION",
    # Errors
    "APIError",
    "AsyncThreeCommon",
    "AuthError",
    "ClientConfig",
    "ConflictError",
    "ConnectionError",
    "NotFoundError",
    "PermissionError",
    "RateLimitError",
    "RetryDelay",
    "ServerError",
    # Clients
    "ThreeCommon",
    "ValidationError",
    "__version__",
)
