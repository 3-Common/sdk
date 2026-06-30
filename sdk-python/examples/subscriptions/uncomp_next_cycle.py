"""Remove a staged comp so the next renewal bills at full price again.

The inverse of ``comp_next_cycle``. A no-op when no comp is pending, and allowed
on a subscription in any state.

Run with: python examples/subscriptions/uncomp_next_cycle.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        sub = client.subscriptions.uncomp_next_cycle("sub_replace_with_real_id")
        print(f"subscription {sub.id} [{sub.status}]")
        print(f"  next renewal ({sub.current_period_end}) will bill at full price")


if __name__ == "__main__":
    main()
