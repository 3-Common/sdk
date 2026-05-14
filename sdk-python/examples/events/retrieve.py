"""Run with: python examples/events/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.events import RetrieveParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        ev = client.events.retrieve(
            "evt_replace_with_real_id",
            RetrieveParams(fields="id,name,start,status"),
        )
        print(f"event {ev.id} — {ev.name!r} [{ev.status}]")
        if ev.start:
            print(f"  starts at {ev.start}")


if __name__ == "__main__":
    main()
