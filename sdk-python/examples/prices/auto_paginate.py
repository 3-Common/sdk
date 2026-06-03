"""Iterate every active price across all products, transparently fetching each
page as the previous one drains.

Run with: python examples/prices/auto_paginate.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.prices import ListParams


def main() -> None:
    count = 0
    recurring = 0
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for price in client.prices.list_auto_paginate(ListParams(active=True)):
            count += 1
            if price.type == "recurring":
                recurring += 1

    print(f"iterated {count} active prices")
    print(f"{recurring} are recurring")


if __name__ == "__main__":
    main()
