"""Run with: python examples/entitlements/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.entitlements import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.entitlements.list(
            ListParams(feature_key="api_calls", min_balance=1, page_size=25)
        )
        print(f"got {len(result.data)} entitlements (has_more={result.has_more})")
        for ent in result.data:
            print(f"  {ent.id} — {ent.contact_id} — balance {ent.balance}")


if __name__ == "__main__":
    asyncio.run(main())
