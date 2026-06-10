"""Resolve a feature's live value for a customer — walks active subscriptions →
prices → feature grants. For quantity features it also reports the current
entitlement balance.

Run with: python examples/features/resolve.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.features import (
    ResolvedFeatureBoolean,
    ResolvedFeatureDuration,
    ResolvedFeatureEnum,
    ResolvedFeatureQuantity,
    ResolveParams,
)


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        resolved = client.features.resolve(
            ResolveParams(contact_id="cnt_replace_with_real_id", feature_key="api_calls")
        )

    value = resolved.value
    print(f"feature {resolved.feature.key} [{value.type}]")
    if isinstance(value, ResolvedFeatureBoolean):
        print(f"  enabled: {value.enabled}")
    elif isinstance(value, ResolvedFeatureQuantity):
        qty = "unlimited" if value.quantity is None else value.quantity
        print(f"  quantity: {qty}")
        if value.balance is not None:
            print(f"  balance:  {value.balance}")
    elif isinstance(value, ResolvedFeatureEnum):
        print(f"  value: {value.enum_value or 'none'}")
    elif isinstance(value, ResolvedFeatureDuration):
        days = "unlimited" if value.duration_days is None else value.duration_days
        print(f"  duration: {days} days")
    print(f"  from subscriptions: {resolved.contributing_subscription_ids}")


if __name__ == "__main__":
    main()
