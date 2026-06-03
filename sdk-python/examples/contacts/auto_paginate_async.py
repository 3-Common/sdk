"""Run with: python examples/contacts/auto_paginate_async.py

Async equivalent of ``auto_paginate.py``. The iterator yields contacts one
page at a time; ``async for`` drains each page transparently.
"""

from __future__ import annotations

import asyncio

from threecommon import AsyncThreeCommon
from threecommon.contacts import ListParams


async def main() -> None:
    total = 0
    last_email = ""
    async with AsyncThreeCommon(api_key="3co_your_api_key_here") as client:
        async for contact in client.contacts.list_auto_paginate(
            ListParams(filter="opted-in"),
        ):
            total += 1
            last_email = contact.email
            if total % 100 == 0:
                print(f"...processed {total} contacts")

    print(f"walked {total} opted-in contacts total (last: {last_email})")


if __name__ == "__main__":
    asyncio.run(main())
