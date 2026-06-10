"""Run with: python examples/prices/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.prices import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.prices.list(
            ListParams(product_id="prod_replace_with_real_id", active=True, page_size=25)
        )
        print(f"got {len(result.data)} prices (has_more={result.has_more})")
        for price in result.data:
            print(f"  {price.id} — {price.type} — {price.unit_amount} {price.currency}")


if __name__ == "__main__":
    asyncio.run(main())
