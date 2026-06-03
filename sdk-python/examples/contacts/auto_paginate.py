"""Run with: python examples/contacts/auto_paginate.py

Walk every opted-in contact for the host with the auto-paginator. Pages are
fetched lazily — one HTTP call per page, only when the previous page's
buffer drains.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.contacts import ListParams


def main() -> None:
    total = 0
    last_email = ""
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for contact in client.contacts.list_auto_paginate(ListParams(filter="opted-in")):
            total += 1
            last_email = contact.email
            if total % 100 == 0:
                print(f"...processed {total} contacts")

    print(f"walked {total} opted-in contacts total (last: {last_email})")


if __name__ == "__main__":
    main()
