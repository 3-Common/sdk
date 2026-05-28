"""Iterate every active subscription, transparently fetching each page as
the previous one drains.

Run with: python examples/subscriptions/auto_paginate.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.subscriptions import ListParams


def main() -> None:
    count = 0
    units = 0
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for sub in client.subscriptions.list_auto_paginate(ListParams(status="active")):
            count += 1
            units += sub.quantity or 0

    print(f"iterated {count} active subscriptions")
    print(f"approximate units in flight: {units}")


if __name__ == "__main__":
    main()
