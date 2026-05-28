"""Create a new subscription with a 14-day trial.

The subscription starts in ``trialing`` and transitions to ``active`` once
the first payment succeeds.

Run with: python examples/subscriptions/create.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.subscriptions import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        sub = client.subscriptions.create(
            CreateBody(
                contact_id="cnt_replace_with_real_id",
                price_id="price_replace_with_real_id",
                quantity=1,
                trial_days=14,
                auto_charge=True,
                notes="Pro plan — annual billing",
                metadata={"source": "website-checkout"},
            )
        )
        print(f"created {sub.id} [{sub.status}]")
        print(f"  trial ends   {sub.trial_end}")
        print(f"  first bill   {sub.current_period_end}")


if __name__ == "__main__":
    main()
