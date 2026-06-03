"""Reactivate a previously archived price. Idempotent.

Run with: python examples/prices/unarchive.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        price = client.prices.unarchive("price_replace_with_real_id")
        print(f"unarchived {price.id} — active={price.active}")


if __name__ == "__main__":
    main()
