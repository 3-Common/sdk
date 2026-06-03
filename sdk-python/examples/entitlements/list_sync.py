"""Run with: python examples/entitlements/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.entitlements import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.entitlements.list(
            ListParams(feature_key="api_calls", min_balance=1, page_size=25)
        )
        print(f"got {len(result.data)} entitlements (has_more={result.has_more})\n")
        for ent in result.data:
            print(f"{ent.id} — {ent.contact_id} — {ent.feature_key} — balance {ent.balance}")


if __name__ == "__main__":
    main()
