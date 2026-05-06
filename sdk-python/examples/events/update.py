"""Run with: python examples/events/update.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.events import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        updated = client.events.update(
            "evt_replace_with_real_id",
            UpdateBody(name="Renamed via SDK"),
        )
        print(f"updated {updated.id} — name is now {updated.name!r}")


if __name__ == "__main__":
    main()
