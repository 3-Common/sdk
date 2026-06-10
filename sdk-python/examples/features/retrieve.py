"""Run with: python examples/features/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        feature = client.features.retrieve("feat_replace_with_real_id")
        print(f"feature {feature.id} [{feature.type}]")
        print(f"  key    {feature.key}")
        print(f"  name   {feature.name}")
        print(f"  active {feature.active}")
        if feature.enum_values is not None:
            print(f"  values {', '.join(feature.enum_values)}")


if __name__ == "__main__":
    main()
