"""Soft-archive a price. Existing subscriptions keep billing; new subscriptions
can no longer select it until unarchived. Idempotent.

Run with: python examples/prices/archive.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        price = client.prices.archive("price_replace_with_real_id")
        print(f"archived {price.id} — active={price.active}")


if __name__ == "__main__":
    main()
