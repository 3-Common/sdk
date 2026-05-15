"""Run with: python examples/invoices/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.invoices import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.invoices.list(
            ListParams(status="open", customer_id="cnt_replace_with_real_id", page_size=25)
        )
        print(f"got {len(result.data)} invoices (has_more={result.has_more})\n")
        for i, inv in enumerate(result.data, start=1):
            print(f"[{i}] {inv.id}")
            print(f"    status:       {inv.status}")
            print(f"    currency:     {inv.currency}")
            print(f"    total:        {inv.total}")
            print(f"    amount_paid:  {inv.amount_paid}")
            print(f"    amount_due:   {inv.amount_due}")
            print()


if __name__ == "__main__":
    main()
