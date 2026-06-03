"""Run with: python examples/contacts/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.contacts import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.contacts.list(
            ListParams(filter="opted-in", page_size=10),
        )
        print(
            f"got {len(result.data)} contacts "
            f"(has_more={result.has_more}, page={result.page_number})",
        )
        for c in result.data:
            print(f"  {c.id} — {c.email} ({c.status})")


if __name__ == "__main__":
    asyncio.run(main())
