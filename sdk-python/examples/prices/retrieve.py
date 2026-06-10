"""Run with: python examples/prices/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        price = client.prices.retrieve("price_replace_with_real_id")
        print(f"price {price.id} [{price.type}]")
        print(f"  product  {price.product_id}")
        print(f"  amount   {price.unit_amount} {price.currency}")
        if price.recurring is not None:
            print(f"  cadence  every {price.recurring.interval_count} {price.recurring.interval}")
        for feature in price.features or []:
            print(f"  feature  {feature.feature_key} [{feature.type}]")


if __name__ == "__main__":
    main()
