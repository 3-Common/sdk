"""Run with: python examples/events/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.events import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.events.list(ListParams(status="open", page_size=10))
        print(f"got {len(result.data)} events (has_more={result.has_more})\n")
        for i, ev in enumerate(result.data, start=1):
            print(f"[{i}] {ev.id}")
            print(f"    name:           {ev.name}")
            print(f"    type:           {ev.type}")
            print(f"    schedule:       {ev.schedule}")
            print(f"    start:          {ev.start}")
            print(f"    status:         {ev.status}")
            print(f"    currency:       {ev.currency}")
            print(f"    items_sold:     {ev.items_sold}")
            print(f"    revenue_cents:  {ev.revenue_cents}")
            print(f"    is_public:      {ev.is_public}")
            print(f"    is_virtual:     {ev.is_virtual}")
            print()


if __name__ == "__main__":
    main()
