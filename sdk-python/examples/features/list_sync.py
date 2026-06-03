"""Run with: python examples/features/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.features import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.features.list(ListParams(type="quantity", active=True, page_size=25))
        print(f"got {len(result.data)} features (has_more={result.has_more})\n")
        for feature in result.data:
            print(f"{feature.id} — {feature.key} — {feature.type}")


if __name__ == "__main__":
    main()
