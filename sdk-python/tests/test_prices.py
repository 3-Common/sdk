"""Prices service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.prices import (
    CreateBody,
    ListParams,
    PriceFeatureQuantity,
    PriceRecurring,
    RetrieveParams,
    UpdateBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


_SAMPLE = {
    "id": "price_123",
    "hostId": "host_1",
    "productId": "prod_7",
    "type": "recurring",
    "currency": "USD",
    "unitAmount": 1500,
    "recurring": {"interval": "month", "intervalCount": 1},
    "features": [
        {"featureKey": "api_calls", "type": "quantity", "quantity": 1000, "rolloverEnabled": False}
    ],
    "nickname": "Pro monthly",
    "active": True,
    "metadata": {},
    "createdAt": "2026-05-01T00:00:00.000Z",
    "updatedAt": "2026-05-01T00:00:00.000Z",
}


# ────────────────────────────────────────────────────────────────────────────
# Sync prices
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices?productId=prod_7&active=true",
        json={"data": [_SAMPLE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.prices.list(ListParams(product_id="prod_7", active=True))
    assert len(result.data) == 1
    price = result.data[0]
    assert price.id == "price_123"
    assert price.recurring is not None
    assert price.recurring.interval == "month"
    assert price.features is not None
    feature = price.features[0]
    assert isinstance(feature, PriceFeatureQuantity)
    assert feature.quantity == 1000
    assert result.has_more is False


def test_list_forwards_lowercase_boolean(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(json={"data": [], "hasMore": False})
    with _make_sync() as c:
        c.prices.list(ListParams(product_id="prod_7", active=False))
    req = httpx_mock.get_requests()[0]
    assert req.url.params.get("active") == "false"
    assert req.url.params.get("productId") == "prod_7"


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/prices", json={"data": [], "hasMore": False})
    with _make_sync() as c:
        result = c.prices.list()
    assert result.data == []


def test_list_empty_params_omits_query(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/prices", json={"data": [], "hasMore": False})
    with _make_sync() as c:
        result = c.prices.list(ListParams())
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/prices/price_123", json={"data": _SAMPLE})
    with _make_sync() as c:
        price = c.prices.retrieve("price_123")
    assert price.id == "price_123"
    assert price.unit_amount == 1500


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_1?fields=id%2CunitAmount",
        json={"data": {"id": "price_1", "unitAmount": 5}},
    )
    with _make_sync() as c:
        price = c.prices.retrieve("price_1", RetrieveParams(fields="id,unitAmount"))
    assert price.unit_amount == 5


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.prices.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.prices.retrieve("price_missing")


def test_create_sends_camelcase_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices",
        method="POST",
        status_code=201,
        match_json={
            "productId": "prod_7",
            "type": "recurring",
            "currency": "USD",
            "unitAmount": 1500,
            "recurring": {"interval": "month", "intervalCount": 1},
            "features": [
                {
                    "featureKey": "api_calls",
                    "type": "quantity",
                    "quantity": 1000,
                    "rolloverEnabled": False,
                }
            ],
        },
        json={"data": _SAMPLE},
    )
    with _make_sync() as c:
        price = c.prices.create(
            CreateBody(
                product_id="prod_7",
                type="recurring",
                currency="USD",
                unit_amount=1500,
                recurring=PriceRecurring(interval="month", interval_count=1),
                features=[
                    PriceFeatureQuantity(
                        feature_key="api_calls",
                        type="quantity",
                        quantity=1000,
                        rollover_enabled=False,
                    )
                ],
            )
        )
    assert price.id == "price_123"


def test_create_preserves_null_quantity(httpx_mock: HTTPXMock) -> None:
    # quantity=None means "unlimited"; it must survive serialization as null.
    httpx_mock.add_response(
        url="http://test.local/v1/prices",
        method="POST",
        status_code=201,
        match_json={
            "productId": "prod_7",
            "type": "recurring",
            "currency": "USD",
            "unitAmount": 0,
            "recurring": {"interval": "month", "intervalCount": 1},
            "features": [
                {
                    "featureKey": "seats",
                    "type": "quantity",
                    "quantity": None,
                    "rolloverEnabled": True,
                }
            ],
        },
        json={"data": _SAMPLE},
    )
    with _make_sync() as c:
        c.prices.create(
            CreateBody(
                product_id="prod_7",
                type="recurring",
                currency="USD",
                unit_amount=0,
                recurring=PriceRecurring(interval="month", interval_count=1),
                features=[
                    PriceFeatureQuantity(
                        feature_key="seats",
                        type="quantity",
                        quantity=None,
                        rollover_enabled=True,
                    )
                ],
            )
        )


def test_create_400_validation(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices",
        method="POST",
        status_code=400,
        json={"error": {"code": "validation_error", "message": "recurring required"}},
    )
    with _make_sync() as c, pytest.raises(ValidationError):
        c.prices.create(
            CreateBody(product_id="prod_7", type="recurring", currency="USD", unit_amount=1500)
        )


def test_create_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.prices.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_update_preserves_explicit_null(httpx_mock: HTTPXMock) -> None:
    # nickname=None clears the field; unset fields (metadata, recurring) are omitted.
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123",
        method="PATCH",
        match_json={"unitAmount": 1200, "nickname": None},
        json={"data": {**_SAMPLE, "unitAmount": 1200, "nickname": None}},
    )
    with _make_sync() as c:
        price = c.prices.update("price_123", UpdateBody(unit_amount=1200, nickname=None))
    assert price.unit_amount == 1200


def test_update_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.prices.update("", UpdateBody(unit_amount=1))
    assert exc.value.code == "missing_id"


def test_update_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.prices.update("price_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_archive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123/archive",
        method="POST",
        json={"data": {**_SAMPLE, "active": False}},
    )
    with _make_sync() as c:
        price = c.prices.archive("price_123")
    assert price.active is False


def test_unarchive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123/unarchive",
        method="POST",
        json={"data": {**_SAMPLE, "active": True}},
    )
    with _make_sync() as c:
        price = c.prices.unarchive("price_123")
    assert price.active is True


def test_archive_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.prices.archive("")
    assert exc.value.code == "missing_id"


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices?page=0&active=true",
        json={"data": [{"id": "price_a"}, {"id": "price_b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/prices?page=1&active=true",
        json={"data": [{"id": "price_c"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [p.id for p in c.prices.list_auto_paginate(ListParams(active=True))]
    assert ids == ["price_a", "price_b", "price_c"]


# ────────────────────────────────────────────────────────────────────────────
# Async prices
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/prices", json={"data": [], "hasMore": False})
    async with _make_async() as c:
        r = await c.prices.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices",
        method="POST",
        status_code=201,
        json={"data": _SAMPLE},
    )
    async with _make_async() as c:
        price = await c.prices.create(
            CreateBody(product_id="prod_7", type="one_time", currency="USD", unit_amount=999)
        )
    assert price.id == "price_123"


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/prices/price_123", json={"data": _SAMPLE})
    async with _make_async() as c:
        price = await c.prices.retrieve("price_123")
    assert price.id == "price_123"
    assert price.unit_amount == 1500


@pytest.mark.asyncio
async def test_async_update(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123",
        method="PATCH",
        match_json={"nickname": None},
        json={"data": {**_SAMPLE, "nickname": None}},
    )
    async with _make_async() as c:
        price = await c.prices.update("price_123", UpdateBody(nickname=None))
    assert price.id == "price_123"


@pytest.mark.asyncio
async def test_async_archive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123/archive",
        method="POST",
        json={"data": {**_SAMPLE, "active": False}},
    )
    async with _make_async() as c:
        price = await c.prices.archive("price_123")
    assert price.active is False


@pytest.mark.asyncio
async def test_async_unarchive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices/price_123/unarchive",
        method="POST",
        json={"data": {**_SAMPLE, "active": True}},
    )
    async with _make_async() as c:
        price = await c.prices.unarchive("price_123")
    assert price.active is True


@pytest.mark.asyncio
async def test_async_create_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.prices.create(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.prices.update("price_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/prices?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/prices?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [p.id async for p in c.prices.list_auto_paginate()]
    assert ids == ["a", "b", "c"]
