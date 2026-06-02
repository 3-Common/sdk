"""Run with: python examples/contacts/list_activity.py

Fetch the activity feed (checkouts, refunds, scans, emails, invoice payments)
for a single contact.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.contacts import ActivityListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.list_activity(
            "cnt_replace_with_real_id",
            ActivityListParams(page_size=20),
        )
        print(f"got {len(result.data)} activity records")
        for event in result.data:
            print(f"  {event.created_at} — {event.type}")


if __name__ == "__main__":
    main()
