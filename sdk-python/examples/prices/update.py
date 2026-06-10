"""Update a price's amount and nickname. To change type, currency, or product,
archive the price and create a new one instead.

Run with: python examples/prices/update.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.prices import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        price = client.prices.update(
            "price_replace_with_real_id",
            UpdateBody(unit_amount=1200, nickname="Pro monthly (promo)"),
        )
        print(f"updated {price.id} — now {price.unit_amount} {price.currency}")


if __name__ == "__main__":
    main()
