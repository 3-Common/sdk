"""Sync and async invoices services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.invoices.types import (
    AutoChargeResult,
    CreateBody,
    DeletedInvoice,
    Invoice,
    ListInvoicesResponse,
    ListParams,
    PaymentBody,
    RefundBody,
    RetrieveParams,
    UpdateBody,
    VoidBody,
)
from threecommon.pagination import AsyncIter, Iter

if TYPE_CHECKING:
    from threecommon._core.http_client import AsyncHTTPClient, HTTPClient


def _encode_list_params(params: ListParams | None) -> dict[str, str] | None:
    if params is None:
        return None
    raw = params.model_dump(by_alias=True, exclude_none=True)
    if not raw:
        return None
    return {k: str(v) for k, v in raw.items()}


def _encode_retrieve_params(params: RetrieveParams | None) -> dict[str, str] | None:
    if params is None or params.fields is None:
        return None
    return {"fields": params.fields}


def _require_id(method: str, invoice_id: str) -> None:
    if not invoice_id:
        msg = f"invoices.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(invoice_id: str) -> str:
    return f"/invoices/{quote(invoice_id, safe='')}"


def _action_path(invoice_id: str, action: str) -> str:
    return f"{_path_for(invoice_id)}/{action}"


def _refund_path(invoice_id: str, payment_id: str) -> str:
    return f"{_path_for(invoice_id)}/payments/{quote(payment_id, safe='')}/refunds"


def _build_auto_charge_result(response: dict[str, Any]) -> AutoChargeResult:
    payload: dict[str, Any] = {"invoice": response["data"], "outcome": response["outcome"]}
    if response.get("failureCode") is not None:
        payload["failure_code"] = response["failureCode"]
    return AutoChargeResult.model_validate(payload)


def _require_payment_id(payment_id: str) -> None:
    if not payment_id:
        msg = "invoices.refund_payment: payment_id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class InvoicesService:
    """Sync invoices service — bound as ``client.invoices`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListInvoicesResponse:
        """List the host's invoices (one page).

        For full iteration use [list_auto_paginate][InvoicesService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/invoices", query=_encode_list_params(params))
        )
        return ListInvoicesResponse.model_validate(body)

    def retrieve(self, invoice_id: str, params: RetrieveParams | None = None) -> Invoice:
        """Retrieve a single invoice by id."""
        _require_id("retrieve", invoice_id)
        body = self._http.request(
            Request(
                method="GET",
                path=_path_for(invoice_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Invoice.model_validate(body["data"])

    def create(self, body: CreateBody) -> Invoice:
        """Create a draft invoice. Totals are computed server-side from line items."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="invoices.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(Request(method="POST", path="/invoices", body=payload))
        return Invoice.model_validate(response["data"])

    def update(self, invoice_id: str, body: UpdateBody) -> Invoice:
        """Revise a draft invoice. Only legal while in draft."""
        _require_id("update", invoice_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="invoices.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(invoice_id), body=payload)
        )
        return Invoice.model_validate(response["data"])

    def finalize(self, invoice_id: str) -> Invoice:
        """Finalize a draft invoice: assign a number, stamp ``issuedAt``, set status ``open``."""
        _require_id("finalize", invoice_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "finalize"), body={})
        )
        return Invoice.model_validate(response["data"])

    def void(self, invoice_id: str, body: VoidBody | None = None) -> Invoice:
        """Void an invoice. Permitted from ``draft`` or ``open``; paid invoices cannot be voided."""
        _require_id("void", invoice_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "void"), body=payload)
        )
        return Invoice.model_validate(response["data"])

    def record_payment(self, invoice_id: str, body: PaymentBody) -> Invoice:
        """Record a manual payment against an open invoice.

        Cumulative payments meeting the total transition the invoice to ``paid``.
        """
        _require_id("record_payment", invoice_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="invoices.record_payment: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "payments"), body=payload)
        )
        return Invoice.model_validate(response["data"])

    def auto_charge(self, invoice_id: str) -> AutoChargeResult:
        """Off-session charge the customer's saved card for an open invoice.

        A decline is not an error — it resolves with ``outcome="failed"`` and a
        ``failure_code``, leaving the invoice in ``payment_failed``. Only
        network / processor 5xx errors raise.
        """
        _require_id("auto_charge", invoice_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "auto_charge"), body={})
        )
        return _build_auto_charge_result(response)

    def refund_payment(self, invoice_id: str, payment_id: str, body: RefundBody) -> Invoice:
        """Refund all or part of a recorded payment on a paid invoice.

        Idempotent on ``body.idempotency_key``: replays return the existing
        refund without contacting the processor again.
        """
        _require_id("refund_payment", invoice_id)
        _require_payment_id(payment_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="invoices.refund_payment: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=_refund_path(invoice_id, payment_id), body=payload)
        )
        return Invoice.model_validate(response["data"])

    def delete_draft(self, invoice_id: str) -> DeletedInvoice:
        """Permanently delete a draft invoice.

        Only legal while in ``draft`` (no number issued); finalized invoices
        must be voided instead so the audit trail stays intact.
        """
        _require_id("delete_draft", invoice_id)
        response = self._http.request(Request(method="DELETE", path=_path_for(invoice_id)))
        return DeletedInvoice.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Invoice]:
        """Iterate every invoice matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Invoice], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/invoices", query=_encode_list_params(page_params))
            )
            response = ListInvoicesResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncInvoicesService:
    """Async invoices service — bound as ``client.invoices`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListInvoicesResponse:
        body = await self._http.request(
            Request(method="GET", path="/invoices", query=_encode_list_params(params))
        )
        return ListInvoicesResponse.model_validate(body)

    async def retrieve(self, invoice_id: str, params: RetrieveParams | None = None) -> Invoice:
        _require_id("retrieve", invoice_id)
        body = await self._http.request(
            Request(
                method="GET",
                path=_path_for(invoice_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Invoice.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Invoice:
        if body is None:
            raise ValidationError(
                code="missing_body", message="invoices.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(Request(method="POST", path="/invoices", body=payload))
        return Invoice.model_validate(response["data"])

    async def update(self, invoice_id: str, body: UpdateBody) -> Invoice:
        _require_id("update", invoice_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="invoices.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(invoice_id), body=payload)
        )
        return Invoice.model_validate(response["data"])

    async def finalize(self, invoice_id: str) -> Invoice:
        _require_id("finalize", invoice_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "finalize"), body={})
        )
        return Invoice.model_validate(response["data"])

    async def void(self, invoice_id: str, body: VoidBody | None = None) -> Invoice:
        _require_id("void", invoice_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = await self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "void"), body=payload)
        )
        return Invoice.model_validate(response["data"])

    async def record_payment(self, invoice_id: str, body: PaymentBody) -> Invoice:
        _require_id("record_payment", invoice_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="invoices.record_payment: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "payments"), body=payload)
        )
        return Invoice.model_validate(response["data"])

    async def auto_charge(self, invoice_id: str) -> AutoChargeResult:
        _require_id("auto_charge", invoice_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(invoice_id, "auto_charge"), body={})
        )
        return _build_auto_charge_result(response)

    async def refund_payment(self, invoice_id: str, payment_id: str, body: RefundBody) -> Invoice:
        _require_id("refund_payment", invoice_id)
        _require_payment_id(payment_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="invoices.refund_payment: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_refund_path(invoice_id, payment_id), body=payload)
        )
        return Invoice.model_validate(response["data"])

    async def delete_draft(self, invoice_id: str) -> DeletedInvoice:
        _require_id("delete_draft", invoice_id)
        response = await self._http.request(Request(method="DELETE", path=_path_for(invoice_id)))
        return DeletedInvoice.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Invoice]:
        """Async iterate every invoice matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Invoice], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/invoices", query=_encode_list_params(page_params))
            )
            response = ListInvoicesResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
