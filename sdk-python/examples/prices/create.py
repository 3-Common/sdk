"""Create a recurring price with a metered feature grant. The quantity grant
refills the customer's entitlement balance on each renewal.

Run with: python examples/prices/create.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.prices import CreateBody, PriceFeatureQuantity, PriceRecurring


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        price = client.prices.create(
            CreateBody(
                product_id="prod_replace_with_real_id",
                type="recurring",
                currency="USD",
                unit_amount=1500,
                recurring=PriceRecurring(interval="month", interval_count=1),
                features=[
                    PriceFeatureQuantity(
                        feature_key="api_calls",
                        type="quantity",
                        quantity=1000,
                        rollover_enabled=False,
                    )
                ],
                nickname="Pro monthly",
                metadata={"tier": "pro"},
            )
        )
        print(f"created {price.id} — {price.unit_amount} {price.currency}")


if __name__ == "__main__":
    main()
