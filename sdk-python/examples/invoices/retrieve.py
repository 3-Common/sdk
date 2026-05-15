"""Run with: python examples/invoices/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        inv = client.invoices.retrieve("inv_replace_with_real_id")
        print(f"invoice {inv.id} [{inv.status}]")
        print(f"  total:       {inv.total} {inv.currency}")
        print(f"  amount_paid: {inv.amount_paid}")
        print(f"  amount_due:  {inv.amount_due}")
        print(f"  line items:  {len(inv.line_items or [])}")
        print(f"  payments:    {len(inv.payments or [])}")


if __name__ == "__main__":
    main()
