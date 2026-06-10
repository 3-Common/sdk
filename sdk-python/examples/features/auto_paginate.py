"""Iterate every active feature in the catalog, transparently fetching each
page as the previous one drains.

Run with: python examples/features/auto_paginate.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.features import ListParams


def main() -> None:
    count = 0
    quantity = 0
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for feature in client.features.list_auto_paginate(ListParams(active=True)):
            count += 1
            if feature.type == "quantity":
                quantity += 1

    print(f"iterated {count} active features")
    print(f"{quantity} are quantity-typed")


if __name__ == "__main__":
    main()
