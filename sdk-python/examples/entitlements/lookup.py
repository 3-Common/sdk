"""Look up the unique entitlement for a (contact, feature) pair.

Raises ``NotFoundError`` if no record exists yet.

Run with: python examples/entitlements/lookup.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.entitlements import LookupParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        ent = client.entitlements.lookup(
            LookupParams(contact_id="cnt_replace_with_real_id", feature_key="api_calls")
        )
        print(f"{ent.contact_id} has {ent.balance} {ent.feature_key} remaining")


if __name__ == "__main__":
    main()
