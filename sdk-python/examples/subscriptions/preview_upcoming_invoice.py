"""Preview the invoice the next renewal will generate (Stripe-style
``invoice.upcoming``).

Returns ``None`` when the subscription is set to cancel at period end.

Run with: python examples/subscriptions/preview_upcoming_invoice.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        preview = client.subscriptions.preview_upcoming_invoice("sub_replace_with_real_id")

    if preview is None:
        print("subscription is set to cancel at period end — no upcoming invoice")
        return

    print(f"next invoice — {preview.total} {preview.currency}")
    print(f"  period {preview.period_start} → {preview.period_end}")
    for line in preview.line_items:
        print(f"  • {line.description} — {line.quantity} x {line.unit_amount}")


if __name__ == "__main__":
    main()
