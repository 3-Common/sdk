"""Run with: python examples/forms/list_async.py"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.forms import ListParams


async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        result = await client.forms.list(ListParams(type="standalone", page_size=10))
        print(f"got {len(result.data)} forms (has_more={result.has_more})")
        for form in result.data:
            print(f"  {form.id} - {form.name} ({form.status})")


if __name__ == "__main__":
    asyncio.run(main())
