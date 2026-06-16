"""Run with: python examples/properties/auto_paginate_async.py

Async equivalent of ``auto_paginate.py``. The iterator yields properties one
page at a time; ``async for`` drains each page transparently.
"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.properties import ListParams


async def main() -> None:
    total = 0
    last_name = ""
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        async for prop in client.properties.list_auto_paginate(
            ListParams(object_type="contact"),
        ):
            total += 1
            last_name = prop.name
            if total % 100 == 0:
                print(f"...processed {total} properties")

    print(f"walked {total} contact properties total (last: {last_name})")


if __name__ == "__main__":
    asyncio.run(main())
