"""Create a draft invoice and finalize it.

Finalizing assigns a sequential number, stamps ``issued_at``, and
transitions the invoice to ``open``.

Run with: python examples/invoices/create_and_finalize.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.invoices import CreateBody, InvoiceLineItem


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        draft = client.invoices.create(
            CreateBody(
                customer_id="cnt_replace_with_real_id",
                currency="USD",
                line_items=[
                    InvoiceLineItem(
                        description="Consulting — May 2026",
                        quantity=8,
                        unit_amount=12_500,
                    ),
                    InvoiceLineItem(
                        description="Onboarding fee",
                        quantity=1,
                        unit_amount=50_000,
                    ),
                ],
                notes="Net 30. Wire transfer preferred.",
            )
        )
        print(f"drafted {draft.id} — total {draft.total} USD")

        issued = client.invoices.finalize(draft.id)
        print(f"finalized {issued.id} as {issued.number} [{issued.status}]")


if __name__ == "__main__":
    main()
