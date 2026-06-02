"""Run with: python examples/contacts/bulk_upsert.py

Bulk-upsert contacts (e.g. from a CSV import). Deduplicated server-side by
email; existing rows are updated rather than rejected.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.contacts import BulkUpsertBody, BulkUpsertItem


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.bulk_upsert(
            BulkUpsertBody(
                contacts=[
                    BulkUpsertItem(email="ada@example.com", first_name="Ada"),
                    BulkUpsertItem(email="beatrix@example.com", first_name="Beatrix"),
                    BulkUpsertItem(email="charles@example.com", first_name="Charles"),
                ]
            )
        )
        print(f"upserted {result.affected} contacts")


if __name__ == "__main__":
    main()
