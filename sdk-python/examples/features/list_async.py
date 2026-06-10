"""Run with: python examples/features/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.features import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.features.list(ListParams(type="quantity", active=True, page_size=25))
        print(f"got {len(result.data)} features (has_more={result.has_more})")
        for feature in result.data:
            print(f"  {feature.id} — {feature.key} — {feature.type}")


if __name__ == "__main__":
    asyncio.run(main())
