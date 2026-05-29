"""Revise a draft invoice.

Only legal while the invoice is in ``draft`` — once finalized, void it and
create a new one instead so the audit trail stays intact. Replacing
``line_items`` recomputes the totals server-side; only the fields you pass are
changed. The method is ``update``; "revise" is the domain term for editing a
draft.

Run with: python examples/invoices/update.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.invoices import InvoiceLineItem, UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        revised = client.invoices.update(
            "inv_replace_with_real_id",
            UpdateBody(
                notes="Net 30. Updated per customer request.",
                due_at="2026-07-01T00:00:00.000Z",
                line_items=[
                    InvoiceLineItem(
                        description="Consulting (revised)",
                        quantity=10,
                        unit_amount=12_500,
                    ),
                ],
            ),
        )
        print(f"revised {revised.id} [{revised.status}]")
        print(f"  subtotal {revised.subtotal}, total {revised.total}")


if __name__ == "__main__":
    main()
