"""Pre-release smoke test against the live API.

Runs a small set of calls and verifies the happy path + the common error
paths across the events, invoices, and subscriptions resources. Used by
.github/workflows/live-smoke.yml (maintainer-only).

Required env:
    THREECOMMON_API_KEY    — a real API key

Optional env:
    THREECOMMON_BASE_URL   — defaults to https://api.3common.com
    SMOKE_EVENT_ID         — an event ID owned by the API-key host; if set,
                             exercises the events.retrieve happy path
    SMOKE_INVOICE_ID       — an invoice ID owned by the API-key host; if set,
                             exercises the invoices.retrieve happy path
    SMOKE_SUBSCRIPTION_ID  — a subscription ID owned by the API-key host; if
                             set, exercises subscriptions.retrieve and
                             subscriptions.preview_upcoming_invoice

Run with: python scripts/livesmoke.py
"""

from __future__ import annotations

import os
import sys
from dataclasses import dataclass

from threecommon import APIError, AuthError, NotFoundError, ThreeCommon
from threecommon.events import ListParams
from threecommon.invoices import ListParams as InvoiceListParams
from threecommon.invoices import RefundBody
from threecommon.subscriptions import ListParams as SubscriptionListParams

# Syntactically valid 24-hex ObjectId that will not match any real record.
# The API rejects non-ObjectId strings with a 400 before reaching the
# not-found path, so this must look well-formed to test 404s.
MISSING_OBJECT_ID = "000000000000000000000000"


@dataclass
class _Result:
    check: str
    status: str  # "pass", "fail", "skip"
    detail: str = ""


def _check_events_list(client: ThreeCommon) -> _Result:
    try:
        result = client.events.list(ListParams(page_size=1))
    except APIError as e:
        return _Result("events.list", "fail", repr(e))
    return _Result(
        "events.list",
        "pass",
        f"data.len={len(result.data)} has_more={result.has_more}",
    )


def _check_events_auto_paginate(client: ThreeCommon) -> _Result:
    try:
        iterator = client.events.list_auto_paginate(ListParams(page_size=1))
        first = next(iterator, None)
    except APIError as e:
        return _Result("events.list_auto_paginate", "fail", repr(e))
    return _Result(
        "events.list_auto_paginate",
        "pass",
        "yielded one" if first is not None else "empty",
    )


def _check_events_retrieve(client: ThreeCommon, known_event_id: str | None) -> _Result:
    if not known_event_id:
        return _Result("events.retrieve", "skip", "SMOKE_EVENT_ID not set")
    try:
        ev = client.events.retrieve(known_event_id)
    except APIError as e:
        return _Result("events.retrieve", "fail", repr(e))
    return _Result("events.retrieve", "pass", f"id={ev.id}")


def _check_events_404(client: ThreeCommon) -> _Result:
    try:
        client.events.retrieve(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("events 404 path", "pass", f"code={e.code} request_id={e.request_id}")
    except APIError as e:
        return _Result("events 404 path", "fail", f"unexpected: {e!r}")
    return _Result("events 404 path", "fail", "expected NotFoundError")


def _check_invoices_list(client: ThreeCommon) -> _Result:
    try:
        result = client.invoices.list(InvoiceListParams(page_size=1))
    except APIError as e:
        return _Result("invoices.list", "fail", repr(e))
    return _Result(
        "invoices.list",
        "pass",
        f"data.len={len(result.data)} has_more={result.has_more}",
    )


def _check_invoices_retrieve(client: ThreeCommon, known_invoice_id: str | None) -> _Result:
    if not known_invoice_id:
        return _Result("invoices.retrieve", "skip", "SMOKE_INVOICE_ID not set")
    try:
        inv = client.invoices.retrieve(known_invoice_id)
    except APIError as e:
        return _Result("invoices.retrieve", "fail", repr(e))
    return _Result("invoices.retrieve", "pass", f"id={inv.id}")


def _check_invoices_404(client: ThreeCommon) -> _Result:
    try:
        client.invoices.retrieve(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("invoices 404 path", "pass", f"code={e.code} request_id={e.request_id}")
    except APIError as e:
        return _Result("invoices 404 path", "fail", f"unexpected: {e!r}")
    return _Result("invoices 404 path", "fail", "expected NotFoundError")


# The new invoice write methods move real money (auto_charge, refund_payment)
# or delete a record (delete_draft), so only their not-found paths are smoke-
# tested against the live host. A well-formed-but-missing id 404s when the
# handler loads the invoice, before any side effect.
def _check_invoices_auto_charge_404(client: ThreeCommon) -> _Result:
    try:
        client.invoices.auto_charge(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("invoices.auto_charge 404 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("invoices.auto_charge 404 path", "fail", f"unexpected: {e!r}")
    return _Result("invoices.auto_charge 404 path", "fail", "expected NotFoundError")


def _check_invoices_refund_404(client: ThreeCommon) -> _Result:
    try:
        client.invoices.refund_payment(MISSING_OBJECT_ID, MISSING_OBJECT_ID, RefundBody(amount=1))
    except NotFoundError as e:
        return _Result("invoices.refund_payment 404 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("invoices.refund_payment 404 path", "fail", f"unexpected: {e!r}")
    return _Result("invoices.refund_payment 404 path", "fail", "expected NotFoundError")


def _check_invoices_delete_draft_404(client: ThreeCommon) -> _Result:
    try:
        client.invoices.delete_draft(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("invoices.delete_draft 404 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("invoices.delete_draft 404 path", "fail", f"unexpected: {e!r}")
    return _Result("invoices.delete_draft 404 path", "fail", "expected NotFoundError")


def _check_subscriptions_list(client: ThreeCommon) -> _Result:
    try:
        result = client.subscriptions.list(SubscriptionListParams(page_size=1))
    except APIError as e:
        return _Result("subscriptions.list", "fail", repr(e))
    return _Result(
        "subscriptions.list",
        "pass",
        f"data.len={len(result.data)} has_more={result.has_more}",
    )


def _check_subscriptions_retrieve(
    client: ThreeCommon, known_subscription_id: str | None
) -> _Result:
    if not known_subscription_id:
        return _Result("subscriptions.retrieve", "skip", "SMOKE_SUBSCRIPTION_ID not set")
    try:
        sub = client.subscriptions.retrieve(known_subscription_id)
    except APIError as e:
        return _Result("subscriptions.retrieve", "fail", repr(e))
    return _Result("subscriptions.retrieve", "pass", f"id={sub.id} status={sub.status}")


def _check_subscriptions_preview(client: ThreeCommon, known_subscription_id: str | None) -> _Result:
    if not known_subscription_id:
        return _Result(
            "subscriptions.preview_upcoming_invoice", "skip", "SMOKE_SUBSCRIPTION_ID not set"
        )
    try:
        preview = client.subscriptions.preview_upcoming_invoice(known_subscription_id)
    except APIError as e:
        return _Result("subscriptions.preview_upcoming_invoice", "fail", repr(e))
    detail = "null (cancel-at-period-end)" if preview is None else f"total={preview.total}"
    return _Result("subscriptions.preview_upcoming_invoice", "pass", detail)


def _check_subscriptions_404(client: ThreeCommon) -> _Result:
    try:
        client.subscriptions.retrieve(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("subscriptions 404 path", "pass", f"code={e.code} request_id={e.request_id}")
    except APIError as e:
        return _Result("subscriptions 404 path", "fail", f"unexpected: {e!r}")
    return _Result("subscriptions 404 path", "fail", "expected NotFoundError")


# The comp/uncomp methods stage and clear a 100%-off next renewal; both mutate
# real subscription state, so only their not-found paths are smoke-tested
# against the live host. A well-formed-but-missing id 404s when the handler
# loads the subscription, before any state change.
def _check_subscriptions_comp_404(client: ThreeCommon) -> _Result:
    try:
        client.subscriptions.comp_next_cycle(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("subscriptions.comp_next_cycle 404 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("subscriptions.comp_next_cycle 404 path", "fail", f"unexpected: {e!r}")
    return _Result("subscriptions.comp_next_cycle 404 path", "fail", "expected NotFoundError")


def _check_subscriptions_uncomp_404(client: ThreeCommon) -> _Result:
    try:
        client.subscriptions.uncomp_next_cycle(MISSING_OBJECT_ID)
    except NotFoundError as e:
        return _Result("subscriptions.uncomp_next_cycle 404 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("subscriptions.uncomp_next_cycle 404 path", "fail", f"unexpected: {e!r}")
    return _Result("subscriptions.uncomp_next_cycle 404 path", "fail", "expected NotFoundError")


def _check_401(base_url: str) -> _Result:
    fake = "3co_smoke_test_invalid_key"  # gitleaks:allow
    try:
        with ThreeCommon(
            api_key=fake,
            base_url=base_url,
            telemetry=False,
            max_retries=0,
        ) as bad:
            bad.events.list(ListParams(page_size=1))
    except AuthError as e:
        return _Result("401 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("401 path", "fail", f"unexpected: {e!r}")
    return _Result("401 path", "fail", "expected AuthError")


def main() -> int:
    api_key = os.environ.get("THREECOMMON_API_KEY", "")
    if not api_key:
        sys.stderr.write("THREECOMMON_API_KEY env var is required for live-smoke runs\n")
        return 1

    base_url = os.environ.get("THREECOMMON_BASE_URL") or "https://api.3common.com"
    known_event_id = os.environ.get("SMOKE_EVENT_ID")
    known_invoice_id = os.environ.get("SMOKE_INVOICE_ID")
    known_subscription_id = os.environ.get("SMOKE_SUBSCRIPTION_ID")

    results: list[_Result] = []
    with ThreeCommon(api_key=api_key, base_url=base_url, telemetry=False) as client:
        results.append(_check_events_list(client))
        results.append(_check_events_auto_paginate(client))
        results.append(_check_events_retrieve(client, known_event_id))
        results.append(_check_events_404(client))
        results.append(_check_invoices_list(client))
        results.append(_check_invoices_retrieve(client, known_invoice_id))
        results.append(_check_invoices_404(client))
        results.append(_check_invoices_auto_charge_404(client))
        results.append(_check_invoices_refund_404(client))
        results.append(_check_invoices_delete_draft_404(client))
        results.append(_check_subscriptions_list(client))
        results.append(_check_subscriptions_retrieve(client, known_subscription_id))
        results.append(_check_subscriptions_preview(client, known_subscription_id))
        results.append(_check_subscriptions_404(client))
        results.append(_check_subscriptions_comp_404(client))
        results.append(_check_subscriptions_uncomp_404(client))
    results.append(_check_401(base_url))

    failed = 0
    for entry in results:
        icon = {"pass": "✓", "fail": "✗", "skip": "○"}.get(entry.status, "?")
        sys.stdout.write(f"{icon} {entry.check} — {entry.detail}\n")
        if entry.status == "fail":
            failed += 1

    if failed:
        sys.stderr.write(f"\n{failed} check(s) failed.\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
