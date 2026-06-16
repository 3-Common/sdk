"""Run with: python examples/forms/auto_paginate.py

Walk every standalone form for the host with the auto-paginator. Pages are
fetched lazily - one HTTP call per page, only when the previous page's
buffer drains.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import ListParams


def main() -> None:
    total = 0
    last_name = ""
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for form in client.forms.list_auto_paginate(ListParams(type="standalone")):
            total += 1
            last_name = form.name
            if total % 100 == 0:
                print(f"...processed {total} forms")

    print(f"walked {total} standalone forms total (last: {last_name})")


if __name__ == "__main__":
    main()
