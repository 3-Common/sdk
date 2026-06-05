"""Debit units from a customer's entitlement balance.

Debits ``one_time_addon`` grants first, then ``manual``, then
``subscription_recurring``. Raises ``ConflictError`` if the balance is
insufficient.

Run with: python examples/entitlements/consume.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.entitlements import ConsumeBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        ent = client.entitlements.consume(
            ConsumeBody(
                contact_id="cnt_replace_with_real_id",
                feature_key="api_calls",
                amount=1,
                reason="POST /v1/generate",
            )
        )
        print(f"consumed 1 — {ent.balance} {ent.feature_key} remaining")


if __name__ == "__main__":
    main()
