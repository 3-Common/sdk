"""Manually grant entitlement units to a customer.

Useful for admin top-ups, comp credits, or migration. Idempotent on
``grant_id``: replaying the same id returns the existing record without
double-crediting.

Run with: python examples/entitlements/grant.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.entitlements import GrantBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        ent = client.entitlements.grant(
            GrantBody(
                contact_id="cnt_replace_with_real_id",
                feature_key="api_calls",
                amount=100,
                grant_id="grant_2026_q2_goodwill",
                metadata={"reason": "service-credit", "approved_by": "ops"},
            )
        )
        print(f"granted — {ent.contact_id} now has {ent.balance} {ent.feature_key}")


if __name__ == "__main__":
    main()
