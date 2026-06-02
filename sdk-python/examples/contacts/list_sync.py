"""Run with: python examples/contacts/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.contacts import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.list(
            ListParams(
                filter="opted-in",
                page_size=50,
                sort_field="mostRecentOrder",
                sort_direction="desc",
            )
        )
        print(
            f"got {len(result.data)} contacts "
            f"(has_more={result.has_more}, page={result.page_number})\n"
        )
        for c in result.data:
            print(f"  {c.id} — {c.email} ({c.status})")


if __name__ == "__main__":
    main()
