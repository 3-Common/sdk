"""Email a customer their invoice.

An ``open`` (or ``payment_failed``) invoice gets a payment-link email; a
``paid`` invoice gets its receipt. ``send`` is re-callable, so it doubles as a
resend — ``draft`` and ``void`` invoices are rejected with a ``409``.

You can also email as part of finalizing by passing ``send_email=True`` to
``finalize``, shown here as the one-step alternative.

Run with: python examples/invoices/send.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        # One step: finalize and email the payment link in a single call.
        issued = client.invoices.finalize("inv_replace_with_real_id", send_email=True)
        print(f"finalized {issued.id} as {issued.number} [{issued.status}]")

        # Or send (or re-send) the email for an already-finalized invoice.
        sent = client.invoices.send("inv_replace_with_real_id")
        print(f"sent invoice {sent.id} [{sent.status}]")


if __name__ == "__main__":
    main()
