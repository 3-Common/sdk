"""Reactivate a previously archived feature. Idempotent.

Run with: python examples/features/unarchive.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        feature = client.features.unarchive("feat_replace_with_real_id")
        print(f"unarchived {feature.id} — active={feature.active}")


if __name__ == "__main__":
    main()
