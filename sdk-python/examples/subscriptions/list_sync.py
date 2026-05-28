"""Run with: python examples/subscriptions/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.subscriptions import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.subscriptions.list(
            ListParams(
                status="active",
                contact_id="cnt_replace_with_real_id",
                page_size=25,
            )
        )
        print(f"got {len(result.data)} subscriptions (has_more={result.has_more})\n")
        for sub in result.data:
            print(f"{sub.id} — {sub.status} — renews {sub.current_period_end}")


if __name__ == "__main__":
    main()
