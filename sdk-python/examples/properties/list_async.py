"""Run with: python examples/properties/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.properties import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.properties.list(
            ListParams(object_type="contact", status="active", page_size=25)
        )
        print(f"got {len(result.data)} properties (has_more={result.has_more})")
        for prop in result.data:
            print(f"  {prop.id} - {prop.type} - {prop.name}")


if __name__ == "__main__":
    asyncio.run(main())
