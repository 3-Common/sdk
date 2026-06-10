"""Create a quantity feature in the catalog. The `key` is the stable
identifier that prices and entitlements reference; `type` decides how the
feature resolves.

Run with: python examples/features/create.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.features import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        feature = client.features.create(
            CreateBody(
                key="api_calls",
                name="API calls",
                type="quantity",
                description="Monthly API call quota",
                metadata={"category": "usage"},
            )
        )
        print(f"created {feature.id} — {feature.key} [{feature.type}]")


if __name__ == "__main__":
    main()
