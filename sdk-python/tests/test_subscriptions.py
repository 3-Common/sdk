"""Subscriptions service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.subscriptions import (
    CancelBody,
    CancelImmediatelyBody,
    CreateBody,
    CreateBodyItem,
    ListParams,
    RetrieveParams,
    UpdateBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


SAMPLE = {
    "id": "sub_123",
    "hostId": "hst_1",
    "contactId": "cnt_42",
    "priceId": "price_1",
    "quantity": 1,
    "status": "active",
    "currentPeriodStart": "2026-01-01T00:00:00Z",
    "currentPeriodEnd": "2026-02-01T00:00:00Z",
    "cancelAtPeriodEnd": False,
    "autoCharge": True,
    "dunningEnabled": True,
}

INVOICE_REF = {"id": "inv_9", "status": "open", "total": 5000, "currency": "USD"}
PRORATION = {"netAmountMinor": 1234, "daysRemaining": 10, "daysInCycle": 30}


# ────────────────────────────────────────────────────────────────────────────
# Sync subscriptions
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions?pageSize=10&status=active",
        json={"data": [SAMPLE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.subscriptions.list(ListParams(status="active", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "sub_123"
    assert result.data[0].status == "active"


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.subscriptions.list()
    assert result.data == []


def test_list_with_default_params_omits_query(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.subscriptions.list(ListParams())
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/subscriptions/sub_123", json={"data": SAMPLE})
    with _make_sync() as c:
        sub = c.subscriptions.retrieve("sub_123")
    assert sub.id == "sub_123"
    assert sub.contact_id == "cnt_42"


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1?fields=id%2Cstatus",
        json={"data": {"id": "sub_1"}},
    )
    with _make_sync() as c:
        c.subscriptions.retrieve("sub_1", RetrieveParams(fields="id,status"))


def test_retrieve_requires_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.subscriptions.retrieve("")
        assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.subscriptions.retrieve("sub_missing")


def test_create_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions",
        method="POST",
        match_json={
            "priceId": "price_1",
            "quantity": 2,
            "contactId": "cnt_42",
            "trialDays": 14,
        },
        json={"data": SAMPLE},
    )
    with _make_sync() as c:
        sub = c.subscriptions.create(
            CreateBody(
                price_id="price_1",
                quantity=2,
                contact_id="cnt_42",
                trial_days=14,
            )
        )
    assert sub.id == "sub_123"


def test_create_sends_multi_item_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions",
        method="POST",
        match_json={
            "items": [{"priceId": "price_a", "quantity": 2}, {"priceId": "price_b"}],
            "contactId": "cnt_42",
        },
        json={"data": SAMPLE},
    )
    with _make_sync() as c:
        c.subscriptions.create(
            CreateBody(
                items=[
                    CreateBodyItem(price_id="price_a", quantity=2),
                    CreateBodyItem(price_id="price_b"),
                ],
                contact_id="cnt_42",
            )
        )


def test_update_returns_proration_and_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123",
        method="PATCH",
        match_json={"priceId": "price_up", "quantity": 2},
        json={
            "data": {**SAMPLE, "priceId": "price_up", "quantity": 2},
            "invoice": INVOICE_REF,
            "proration": PRORATION,
        },
    )
    with _make_sync() as c:
        result = c.subscriptions.update("sub_123", UpdateBody(price_id="price_up", quantity=2))
    assert result.subscription.price_id == "price_up"
    assert result.invoice is not None
    assert result.invoice.id == "inv_9"
    assert result.proration.net_amount_minor == 1234
    assert result.proration.days_remaining == 10


def test_update_handles_no_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123",
        method="PATCH",
        json={"data": SAMPLE, "proration": PRORATION},
    )
    with _make_sync() as c:
        result = c.subscriptions.update("sub_123", UpdateBody(quantity=1))
    assert result.invoice is None
    assert result.proration.days_in_cycle == 30


def test_update_validates_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.subscriptions.update("", UpdateBody(quantity=1))
        assert exc.value.code == "missing_id"


def test_activate_posts(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/activate",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "status": "active"}},
    )
    with _make_sync() as c:
        sub = c.subscriptions.activate("sub_123")
    assert sub.status == "active"


def test_activate_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.subscriptions.activate("")


def test_cancel_with_reason(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/cancel",
        method="POST",
        match_json={"reason": "churn"},
        json={"data": {**SAMPLE, "cancelAtPeriodEnd": True}},
    )
    with _make_sync() as c:
        sub = c.subscriptions.cancel("sub_123", CancelBody(reason="churn"))
    assert sub.cancel_at_period_end is True


def test_cancel_without_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/cancel",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "cancelAtPeriodEnd": True}},
    )
    with _make_sync() as c:
        sub = c.subscriptions.cancel("sub_123")
    assert sub.cancel_at_period_end is True


def test_cancel_immediately(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/cancel-immediately",
        method="POST",
        match_json={"reason": "fraud"},
        json={"data": {**SAMPLE, "status": "canceled", "endedAt": "2026-01-15T00:00:00Z"}},
    )
    with _make_sync() as c:
        sub = c.subscriptions.cancel_immediately("sub_123", CancelImmediatelyBody(reason="fraud"))
    assert sub.status == "canceled"
    assert sub.ended_at == "2026-01-15T00:00:00Z"


def test_mark_unpaid(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/mark-unpaid",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "status": "unpaid"}},
    )
    with _make_sync() as c:
        sub = c.subscriptions.mark_unpaid("sub_123")
    assert sub.status == "unpaid"


def test_bill_returns_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/bill",
        method="POST",
        match_json={},
        json={"data": SAMPLE, "invoice": INVOICE_REF},
    )
    with _make_sync() as c:
        result = c.subscriptions.bill("sub_123")
    assert result.subscription.id == "sub_123"
    assert result.invoice.id == "inv_9"
    assert result.invoice.total == 5000


def test_renew_with_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/renew",
        method="POST",
        match_json={},
        json={"data": SAMPLE, "invoice": INVOICE_REF},
    )
    with _make_sync() as c:
        result = c.subscriptions.renew("sub_123")
    assert result.invoice is not None
    assert result.invoice.id == "inv_9"


def test_renew_without_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/renew",
        method="POST",
        json={"data": {**SAMPLE, "status": "canceled"}},
    )
    with _make_sync() as c:
        result = c.subscriptions.renew("sub_123")
    assert result.invoice is None
    assert result.subscription.status == "canceled"


def test_preview_upcoming_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/upcoming",
        json={
            "data": {
                "invoice": {
                    "customerId": "cnt_42",
                    "subscriptionId": "sub_123",
                    "currency": "USD",
                    "lineItems": [
                        {
                            "description": "Pro plan",
                            "quantity": 1,
                            "unitAmount": 5000,
                        }
                    ],
                    "subtotal": 5000,
                    "total": 5000,
                    "periodStart": "2026-02-01T00:00:00Z",
                    "periodEnd": "2026-03-01T00:00:00Z",
                }
            }
        },
    )
    with _make_sync() as c:
        preview = c.subscriptions.preview_upcoming_invoice("sub_123")
    assert preview is not None
    assert preview.total == 5000
    assert preview.line_items[0].unit_amount == 5000


def test_preview_upcoming_invoice_null(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_123/upcoming",
        json={"data": {"invoice": None}},
    )
    with _make_sync() as c:
        preview = c.subscriptions.preview_upcoming_invoice("sub_123")
    assert preview is None


def test_preview_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.subscriptions.preview_upcoming_invoice("")


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions?page=0&status=active",
        json={"data": [{"id": "sub_1"}, {"id": "sub_2"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions?page=1&status=active",
        json={"data": [{"id": "sub_3"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [s.id for s in c.subscriptions.list_auto_paginate(ListParams(status="active"))]
    assert ids == ["sub_1", "sub_2", "sub_3"]


# ────────────────────────────────────────────────────────────────────────────
# Async subscriptions
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions", json={"data": [], "hasMore": False}
    )
    async with _make_async() as c:
        r = await c.subscriptions.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1", json={"data": {**SAMPLE, "id": "sub_1"}}
    )
    async with _make_async() as c:
        sub = await c.subscriptions.retrieve("sub_1")
    assert sub.id == "sub_1"


@pytest.mark.asyncio
async def test_async_update_returns_proration(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1",
        method="PATCH",
        match_json={"quantity": 3},
        json={"data": {**SAMPLE, "id": "sub_1", "quantity": 3}, "proration": PRORATION},
    )
    async with _make_async() as c:
        result = await c.subscriptions.update("sub_1", UpdateBody(quantity=3))
    assert result.subscription.quantity == 3
    assert result.invoice is None
    assert result.proration.net_amount_minor == 1234


@pytest.mark.asyncio
async def test_async_cancel_default_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1/cancel",
        method="POST",
        match_json={},
        json={"data": {**SAMPLE, "id": "sub_1", "cancelAtPeriodEnd": True}},
    )
    async with _make_async() as c:
        sub = await c.subscriptions.cancel("sub_1")
    assert sub.cancel_at_period_end is True


@pytest.mark.asyncio
async def test_async_bill(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1/bill",
        method="POST",
        json={"data": {**SAMPLE, "id": "sub_1"}, "invoice": INVOICE_REF},
    )
    async with _make_async() as c:
        result = await c.subscriptions.bill("sub_1")
    assert result.invoice.id == "inv_9"


@pytest.mark.asyncio
async def test_async_renew_no_invoice(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1/renew",
        method="POST",
        json={"data": {**SAMPLE, "id": "sub_1", "status": "canceled"}},
    )
    async with _make_async() as c:
        result = await c.subscriptions.renew("sub_1")
    assert result.invoice is None
    assert result.subscription.status == "canceled"


@pytest.mark.asyncio
async def test_async_preview_upcoming_null(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions/sub_1/upcoming",
        json={"data": {"invoice": None}},
    )
    async with _make_async() as c:
        preview = await c.subscriptions.preview_upcoming_invoice("sub_1")
    assert preview is None


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/subscriptions?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [s.id async for s in c.subscriptions.list_auto_paginate()]
    assert ids == ["a", "b", "c"]


# ────────────────────────────────────────────────────────────────────────────
# Body-validation guards
# ────────────────────────────────────────────────────────────────────────────


def test_create_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.subscriptions.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_update_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.subscriptions.update("sub_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_create_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.subscriptions.create(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.subscriptions.update("sub_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"
