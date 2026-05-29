"""Off-session auto-charge an open invoice against the customer's saved card.

A decline is not an error — the call resolves with ``outcome == "failed"`` and
a ``failure_code``, leaving the invoice in ``payment_failed``. Only network /
processor (5xx) errors raise.

Run with: python examples/invoices/auto_charge.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.invoices.auto_charge("inv_replace_with_real_id")
        if result.outcome == "paid":
            print(f"invoice {result.invoice.id} charged, now {result.invoice.status}")
        else:
            print(
                f"charge failed ({result.failure_code or 'unknown'}); "
                f"invoice is {result.invoice.status}"
            )


if __name__ == "__main__":
    main()
