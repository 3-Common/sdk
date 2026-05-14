"""Run with: python examples/events/filters_demo.py"""

from __future__ import annotations

from threecommon import ThreeCommon, filters
from threecommon.events import ListParams


def main() -> None:
    f = filters.and_(
        filters.field("status").is_any_of(["open"]),
        filters.field("ticket_sum").is_greater_than(10),
        filters.or_(
            filters.field("type").is_equal_to("event"),
            filters.field("type").is_equal_to("class"),
        ),
    )

    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.events.list(ListParams(filters=f.serialize()))
        print(f"matched {len(result.data)} events")
        for ev in result.data:
            print(f"  {ev.id} — {ev.name} [tickets sold: {ev.items_sold or 0}]")


if __name__ == "__main__":
    main()
