"""Apply a mid-cycle upgrade.

Returns the updated subscription, a proration summary, and (when the rate
difference is positive) a slim reference to the proration invoice.

Run with: python examples/subscriptions/update.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.subscriptions import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.subscriptions.update(
            "sub_replace_with_real_id",
            UpdateBody(
                price_id="price_upgrade_replace_with_real_id",
                quantity=2,
            ),
        )
        sub = result.subscription
        proration = result.proration
        print(f"updated {sub.id} → {sub.price_id} x {sub.quantity}")
        print(
            f"proration: {proration.net_amount_minor} minor units "
            f"({proration.days_remaining}/{proration.days_in_cycle} days)"
        )
        if result.invoice is not None:
            inv = result.invoice
            print(f"proration invoice {inv.id} [{inv.status}] — total {inv.total} {inv.currency}")
        else:
            print("downgrade or no-op — no proration invoice issued")


if __name__ == "__main__":
    main()
