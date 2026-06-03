"""Run with: python examples/contacts/filters_demo.py

Build a typed filter for the contacts list. The ``filters`` namespace is
shared across resources — every endpoint that accepts ``filters`` consumes
the same builder.

The simple ``filter`` enum (``opted-in``, ``unknown``, ...) and the rich
``filters`` builder can be combined; the server ANDs them.
"""

from __future__ import annotations

from threecommon import ThreeCommon, filters
from threecommon.contacts import ListParams


def main() -> None:
    # High-value opted-in contacts whose most recent order is in 2026.
    f = filters.and_(
        filters.field("status").is_any_of(["opted-in"]),
        filters.field("grossSum").is_greater_than(100_000),
        filters.or_(
            filters.field("orderSum").is_greater_than_or_equal_to(5),
            filters.field("lastOrder").is_after("2026-01-01T00:00:00.000Z"),
        ),
    )

    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.list(
            ListParams(
                filters=f.serialize(),
                sort_field="grossSum",
                sort_direction="desc",
                page_size=25,
            )
        )
        print(f"matched {len(result.data)} contacts (has_more={result.has_more})")
        for c in result.data:
            print(f"  {c.full_name} <{c.email}> — gross {c.gross_sum}")


if __name__ == "__main__":
    main()
