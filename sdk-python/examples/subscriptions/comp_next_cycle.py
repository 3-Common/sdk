"""Stage a one-time fully-free (100% off) next renewal cycle.

The next renewal consumes the comp exactly once, then billing resumes at full
price. Rejected on a ``canceled`` or ``unpaid`` subscription.

Run with: python examples/subscriptions/comp_next_cycle.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        sub = client.subscriptions.comp_next_cycle("sub_replace_with_real_id")
        print(f"subscription {sub.id} [{sub.status}]")
        print(f"  next renewal ({sub.current_period_end}) will be comped")


if __name__ == "__main__":
    main()
