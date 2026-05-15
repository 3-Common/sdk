"""Iterate every open invoice for a customer and sum the amounts due.

Pages are fetched lazily — one HTTP call per page, only when the previous
page's buffer drains.

Run with: python examples/invoices/auto_paginate.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.invoices import ListParams


def main() -> None:
    total_due = 0
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        for invoice in client.invoices.list_auto_paginate(
            ListParams(status="open", customer_id="cnt_replace_with_real_id")
        ):
            total_due += invoice.amount_due or 0

    print(f"total amount due across all open invoices: {total_due} cents")


if __name__ == "__main__":
    main()
