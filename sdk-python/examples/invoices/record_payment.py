"""Record a manual payment against an open invoice.

The ``idempotency_key`` makes the request safe to replay — recording the same
payment twice with the same key is a no-op.

Run with: python examples/invoices/record_payment.py
"""

from __future__ import annotations

from datetime import datetime, timezone

from threecommon import ThreeCommon
from threecommon.invoices import PaymentBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        updated = client.invoices.record_payment(
            "inv_replace_with_real_id",
            PaymentBody(
                payment=50_000,  # $500.00 in cents
                idempotency_key=f"pmt-{datetime.now(timezone.utc).isoformat()}",
                note="Wire transfer, ref ABCD-1234",
            ),
        )
        print(f"invoice {updated.id} now {updated.status}")
        print(f"  paid: {updated.amount_paid}, due: {updated.amount_due}")


if __name__ == "__main__":
    main()
