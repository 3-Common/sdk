"""Run with: python examples/events/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.events import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.events.list(ListParams(status="open", page_size=10))
        print(f"got {len(result.data)} events (has_more={result.has_more})")
        for ev in result.data:
            print(f"  {ev.id} — {ev.name} [{ev.status}]")


if __name__ == "__main__":
    asyncio.run(main())
