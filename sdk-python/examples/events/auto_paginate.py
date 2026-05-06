"""Run with: python examples/events/auto_paginate.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.events import ListParams


def sync_main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        count = 0
        for ev in client.events.list_auto_paginate(ListParams(status="open")):
            count += 1
            print(f"{count:4d}. {ev.id} — {ev.name}")
        print(f"walked {count} events total")


async def async_main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        count = 0
        async for ev in client.events.list_auto_paginate(ListParams(status="open")):
            count += 1
            print(f"{count:4d}. {ev.id} — {ev.name}")
        print(f"walked {count} events total")


if __name__ == "__main__":
    sync_main()
    print("---async---")
    asyncio.run(async_main())
