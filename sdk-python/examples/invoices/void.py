"""Void an invoice.

Permitted from ``draft`` or ``open``. Paid invoices cannot be voided —
issue a credit note or refund the payment instead.

Run with: python examples/invoices/void.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.invoices import VoidBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        voided = client.invoices.void(
            "inv_replace_with_real_id",
            VoidBody(reason="Sent to the wrong customer"),
        )
        print(f"invoice {voided.id} status: {voided.status}")


if __name__ == "__main__":
    main()
