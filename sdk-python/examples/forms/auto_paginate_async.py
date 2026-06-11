"""Run with: python examples/forms/auto_paginate_async.py

Async equivalent of ``auto_paginate.py``. The iterator yields forms one page
at a time; ``async for`` drains each page transparently.
"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.forms import ListParams


async def main() -> None:
    total = 0
    last_name = ""
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        async for form in client.forms.list_auto_paginate(ListParams(type="standalone")):
            total += 1
            last_name = form.name
            if total % 100 == 0:
                print(f"...processed {total} forms")

    print(f"walked {total} standalone forms total (last: {last_name})")


if __name__ == "__main__":
    asyncio.run(main())
