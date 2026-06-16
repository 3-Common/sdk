"""Run with: python examples/properties/auto_paginate.py

Walk every contact property for the host with the auto-paginator. Pages are
fetched lazily - one HTTP call per page, only when the previous page's
buffer drains.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.properties import ListParams


def main() -> None:
    total = 0
    last_name = ""
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for prop in client.properties.list_auto_paginate(ListParams(object_type="contact")):
            total += 1
            last_name = prop.name
            if total % 100 == 0:
                print(f"...processed {total} properties")

    print(f"walked {total} contact properties total (last: {last_name})")


if __name__ == "__main__":
    main()
