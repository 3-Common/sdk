"""Run with: python examples/properties/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        prop = client.properties.retrieve("prop_replace_with_real_id")
        print(f"property {prop.id} [{prop.type}]")
        print(f"  name        {prop.name}")
        print(f"  object_type {prop.object_type}")
        print(f"  status      {prop.status}")
        for option in prop.options or []:
            print(f"  option      {option.value} -> {option.label}")


if __name__ == "__main__":
    main()
