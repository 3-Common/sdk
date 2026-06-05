"""Iterate every entitlement for a feature, transparently fetching each page
as the previous one drains. Handy for usage reports or sweeping for
low-balance customers.

Run with: python examples/entitlements/auto_paginate.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.entitlements import ListParams

LOW_BALANCE_THRESHOLD = 10


def main() -> None:
    count = 0
    low_balance = 0
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for ent in client.entitlements.list_auto_paginate(ListParams(feature_key="api_calls")):
            count += 1
            if (ent.balance or 0) < LOW_BALANCE_THRESHOLD:
                low_balance += 1

    print(f"iterated {count} entitlements")
    print(f"{low_balance} are running low (balance < {LOW_BALANCE_THRESHOLD})")


if __name__ == "__main__":
    main()
