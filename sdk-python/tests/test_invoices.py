"""Invoices service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.invoices import (
    CreateBody,
    InvoiceLineItem,
    ListParams,
    PaymentBody,
    RefundBody,
    RetrieveParams,
    UpdateBody,
    VoidBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


SAMPLE = {
    "id": "inv_123",
    "hostId": "hst_1",
    "customerId": "cnt_42",
    "currency": "USD",
    "status": "draft",
    "total": 50000,
    "amountDue": 50000,
    "amountPaid": 0,
}


# ────────────────────────────────────────────────────────────────────────────
# Sync invoices
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?pageSize=10&status=open",
        json={"data": [SAMPLE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.invoices.list(ListParams(status="open", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "inv_123"
    assert result.data[0].status == "draft"


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.invoices.list()
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/invoices/inv_123", json={"data": SAMPLE})
    with _make_sync() as c:
        inv = c.invoices.retrieve("inv_123")
    assert inv.id == "inv_123"
    assert inv.customer_id == "cnt_42"


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1?fields=id%2Cstatus",
        json={"data": {"id": "inv_1"}},
    )
    with _make_sync() as c:
        c.invoices.retrieve("inv_1", RetrieveParams(fields="id,status"))


def test_retrieve_requires_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.invoices.retrieve("")
        assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.invoices.retrieve("inv_missing")


def test_create_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices",
        method="POST",
        match_json={
            "customerId": "cnt_42",
            "currency": "USD",
            "lineItems": [{"description": "Consulting", "quantity": 1, "unitAmount": 50000}],
        },
        json={"data": SAMPLE},
    )
    with _make_sync() as c:
        inv = c.invoices.create(
            CreateBody(
                customer_id="cnt_42",
                currency="USD",
                line_items=[
                    InvoiceLineItem(description="Consulting", quantity=1, unit_amount=50000)
                ],
            )
        )
    assert inv.id == "inv_123"


def test_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123",
        method="PATCH",
        match_json={"notes": "Net 30"},
        json={"data": {**SAMPLE, "notes": "Net 30"}},
    )
    with _make_sync() as c:
        inv = c.invoices.update("inv_123", UpdateBody(notes="Net 30"))
    assert inv.notes == "Net 30"


def test_update_validates_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.invoices.update("", UpdateBody(notes="x"))
        assert exc.value.code == "missing_id"


def test_finalize_posts(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/finalize",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "status": "open", "number": "INV-0001"}},
    )
    with _make_sync() as c:
        inv = c.invoices.finalize("inv_123")
    assert inv.status == "open"
    assert inv.number == "INV-0001"


def test_finalize_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.invoices.finalize("")


def test_void_with_reason(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/void",
        method="POST",
        match_json={"reason": "Sent in error"},
        json={"data": {**SAMPLE, "status": "void"}},
    )
    with _make_sync() as c:
        inv = c.invoices.void("inv_123", VoidBody(reason="Sent in error"))
    assert inv.status == "void"


def test_void_without_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/void",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "status": "void"}},
    )
    with _make_sync() as c:
        inv = c.invoices.void("inv_123")
    assert inv.status == "void"


def test_record_payment_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/payments",
        method="POST",
        match_json={"payment": 50000, "idempotencyKey": "pmt-1"},
        json={"data": {**SAMPLE, "status": "paid", "amountPaid": 50000, "amountDue": 0}},
    )
    with _make_sync() as c:
        inv = c.invoices.record_payment(
            "inv_123", PaymentBody(payment=50000, idempotency_key="pmt-1")
        )
    assert inv.status == "paid"
    assert inv.amount_due == 0


def test_record_payment_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.invoices.record_payment("", PaymentBody(payment=1))


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?page=0&status=open",
        json={"data": [{"id": "inv_1"}, {"id": "inv_2"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?page=1&status=open",
        json={"data": [{"id": "inv_3"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [inv.id for inv in c.invoices.list_auto_paginate(ListParams(status="open"))]
    assert ids == ["inv_1", "inv_2", "inv_3"]


# ────────────────────────────────────────────────────────────────────────────
# Async invoices
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices", json={"data": [], "hasMore": False}
    )
    async with _make_async() as c:
        r = await c.invoices.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_finalize(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1/finalize",
        method="POST",
        json={"data": {**SAMPLE, "id": "inv_1", "status": "open"}},
    )
    async with _make_async() as c:
        inv = await c.invoices.finalize("inv_1")
    assert inv.status == "open"


@pytest.mark.asyncio
async def test_async_void_default_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1/void",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "id": "inv_1", "status": "void"}},
    )
    async with _make_async() as c:
        inv = await c.invoices.void("inv_1")
    assert inv.status == "void"


@pytest.mark.asyncio
async def test_async_record_payment(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1/payments",
        method="POST",
        match_json={"payment": 100},
        json={"data": {**SAMPLE, "id": "inv_1", "amountPaid": 100}},
    )
    async with _make_async() as c:
        inv = await c.invoices.record_payment("inv_1", PaymentBody(payment=100))
    assert inv.amount_paid == 100


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [inv.id async for inv in c.invoices.list_auto_paginate()]
    assert ids == ["a", "b", "c"]


# ────────────────────────────────────────────────────────────────────────────
# Body-validation guards
# ────────────────────────────────────────────────────────────────────────────


def test_create_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_update_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.update("inv_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_record_payment_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.record_payment("inv_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_create_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.invoices.create(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.invoices.update("inv_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1",
        method="PATCH",
        match_json={"notes": "Net 30"},
        json={"data": {**SAMPLE, "id": "inv_1", "notes": "Net 30"}},
    )
    async with _make_async() as c:
        inv = await c.invoices.update("inv_1", UpdateBody(notes="Net 30"))
    assert inv.notes == "Net 30"


@pytest.mark.asyncio
async def test_async_record_payment_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.invoices.record_payment("inv_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


def test_list_with_default_params_omits_query(httpx_mock: HTTPXMock) -> None:
    """ListParams() with no fields set must not produce a query string."""
    httpx_mock.add_response(
        url="http://test.local/v1/invoices",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.invoices.list(ListParams())
    assert result.data == []


def test_list_forwards_subscription_id(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices?subscriptionId=sub_99",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.invoices.list(ListParams(subscription_id="sub_99"))
    assert result.data == []


# ────────────────────────────────────────────────────────────────────────────
# auto_charge / refund_payment / delete_draft
# ────────────────────────────────────────────────────────────────────────────


def test_auto_charge_paid(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/auto_charge",
        method="POST",
        match_json={},
        json={
            "data": {**SAMPLE, "status": "paid", "amountPaid": 50000, "amountDue": 0},
            "outcome": "paid",
        },
    )
    with _make_sync() as c:
        result = c.invoices.auto_charge("inv_123")
    assert result.outcome == "paid"
    assert result.invoice.status == "paid"
    assert result.failure_code is None


def test_auto_charge_declined(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/auto_charge",
        method="POST",
        json={
            "data": {**SAMPLE, "status": "payment_failed"},
            "outcome": "failed",
            "failureCode": "card_declined",
        },
    )
    with _make_sync() as c:
        result = c.invoices.auto_charge("inv_123")
    assert result.outcome == "failed"
    assert result.invoice.status == "payment_failed"
    assert result.failure_code == "card_declined"


def test_auto_charge_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.auto_charge("")
    assert exc.value.code == "missing_id"


def test_refund_payment_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123/payments/pay_456/refunds",
        method="POST",
        match_json={
            "amount": 25000,
            "reason": "requested_by_customer",
            "idempotencyKey": "rfnd-1",
        },
        json={"data": {**SAMPLE, "status": "paid"}},
    )
    with _make_sync() as c:
        inv = c.invoices.refund_payment(
            "inv_123",
            "pay_456",
            RefundBody(amount=25000, reason="requested_by_customer", idempotency_key="rfnd-1"),
        )
    assert inv.id == "inv_123"
    assert inv.status == "paid"


def test_refund_payment_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.refund_payment("", "pay_456", RefundBody(amount=1))
    assert exc.value.code == "missing_id"


def test_refund_payment_requires_payment_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.refund_payment("inv_123", "", RefundBody(amount=1))
    assert exc.value.code == "missing_id"


def test_refund_payment_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.refund_payment("inv_1", "pay_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_delete_draft(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_123",
        method="DELETE",
        json={"data": {"id": "inv_123"}},
    )
    with _make_sync() as c:
        result = c.invoices.delete_draft("inv_123")
    assert result.id == "inv_123"


def test_delete_draft_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.invoices.delete_draft("")
    assert exc.value.code == "missing_id"


@pytest.mark.asyncio
async def test_async_auto_charge(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1/auto_charge",
        method="POST",
        json={"data": {**SAMPLE, "id": "inv_1", "status": "paid"}, "outcome": "paid"},
    )
    async with _make_async() as c:
        result = await c.invoices.auto_charge("inv_1")
    assert result.outcome == "paid"
    assert result.invoice.status == "paid"


@pytest.mark.asyncio
async def test_async_refund_payment(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1/payments/pay_9/refunds",
        method="POST",
        match_json={"amount": 1000},
        json={"data": {**SAMPLE, "id": "inv_1", "status": "paid"}},
    )
    async with _make_async() as c:
        inv = await c.invoices.refund_payment("inv_1", "pay_9", RefundBody(amount=1000))
    assert inv.status == "paid"


@pytest.mark.asyncio
async def test_async_delete_draft(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/invoices/inv_1",
        method="DELETE",
        json={"data": {"id": "inv_1"}},
    )
    async with _make_async() as c:
        result = await c.invoices.delete_draft("inv_1")
    assert result.id == "inv_1"


@pytest.mark.asyncio
async def test_async_refund_payment_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.invoices.refund_payment("inv_1", "pay_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"
