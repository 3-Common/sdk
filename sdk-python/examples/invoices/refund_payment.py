"""Refund all or part of a recorded payment on a paid invoice.

The ``idempotency_key`` makes the request safe to replay — refunding twice with
the same key returns the existing refund instead of issuing a second one.

Run with: python examples/invoices/refund_payment.py
"""

from __future__ import annotations

from datetime import datetime, timezone

from threecommon import ThreeCommon
from threecommon.invoices import RefundBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        refunded = client.invoices.refund_payment(
            "inv_replace_with_real_id",
            "pay_replace_with_real_id",
            RefundBody(
                amount=25_000,  # $250.00 in cents; capped at the refundable balance
                reason="requested_by_customer",
                idempotency_key=f"rfnd-{datetime.now(timezone.utc).isoformat()}",
            ),
        )
        print(f"invoice {refunded.id} now {refunded.status}")
        print(f"  paid: {refunded.amount_paid}, due: {refunded.amount_due}")


if __name__ == "__main__":
    main()
