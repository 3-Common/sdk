"""Run with: python examples/properties/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.properties import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.properties.list(
            ListParams(object_type="contact", status="active", page_size=25)
        )
        print(f"got {len(result.data)} properties (has_more={result.has_more})\n")
        for prop in result.data:
            print(f"{prop.id} - {prop.type} - {prop.name}")


if __name__ == "__main__":
    main()
