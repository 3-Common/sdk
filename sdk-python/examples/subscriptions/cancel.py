"""Schedule cancellation at the end of the current period.

The customer retains access until ``current_period_end``; the next renewal
transitions the subscription to ``canceled`` instead of advancing.

Run with: python examples/subscriptions/cancel.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.subscriptions import CancelBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        sub = client.subscriptions.cancel(
            "sub_replace_with_real_id",
            CancelBody(reason="Customer requested via support ticket #4821"),
        )
        print(f"subscription {sub.id} [{sub.status}]")
        print(f"  cancelAtPeriodEnd: {sub.cancel_at_period_end}")
        print(f"  access continues until {sub.current_period_end}")


if __name__ == "__main__":
    main()
