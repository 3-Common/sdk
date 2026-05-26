"""Run with: python examples/subscriptions/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        sub = client.subscriptions.retrieve("sub_replace_with_real_id")
        print(f"subscription {sub.id} [{sub.status}]")
        print(f"  price          {sub.price_id} x {sub.quantity}")
        print(f"  current period {sub.current_period_start} → {sub.current_period_end}")
        print(f"  cancelAtPeriodEnd: {sub.cancel_at_period_end}")
        print(f"  autoCharge: {sub.auto_charge}")


if __name__ == "__main__":
    main()
