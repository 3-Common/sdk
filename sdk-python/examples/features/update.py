"""Update a feature's display fields. `key` and `type` are immutable — archive
and create a new feature to change them.

Run with: python examples/features/update.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.features import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        feature = client.features.update(
            "feat_replace_with_real_id",
            UpdateBody(name="API requests", description="Monthly API request quota"),
        )
        print(f"updated {feature.id} — {feature.name}")


if __name__ == "__main__":
    main()
