"""Features service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    ConflictError,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.features import (
    CreateBody,
    ListParams,
    ResolvedFeatureQuantity,
    ResolveParams,
    RetrieveParams,
    UpdateBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


_SAMPLE = {
    "id": "feat_123",
    "hostId": "host_1",
    "key": "api_calls",
    "name": "API calls",
    "description": "Monthly API call quota",
    "type": "quantity",
    "active": True,
    "metadata": {},
    "createdAt": "2026-05-01T00:00:00.000Z",
    "updatedAt": "2026-05-01T00:00:00.000Z",
}

_RESOLVED = {
    "feature": _SAMPLE,
    "value": {"type": "quantity", "quantity": 1000, "balance": 850},
    "contributingSubscriptionIds": ["sub_1"],
}


# ────────────────────────────────────────────────────────────────────────────
# Sync features
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features?type=quantity&active=true",
        json={"data": [_SAMPLE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.features.list(ListParams(type="quantity", active=True))
    assert len(result.data) == 1
    assert result.data[0].key == "api_calls"
    assert result.data[0].type == "quantity"
    assert result.has_more is False


def test_list_forwards_lowercase_boolean(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(json={"data": [], "hasMore": False})
    with _make_sync() as c:
        c.features.list(ListParams(type="enum", active=False))
    req = httpx_mock.get_requests()[0]
    assert req.url.params.get("active") == "false"
    assert req.url.params.get("type") == "enum"


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.features.list()
    assert result.data == []


def test_list_empty_params_omits_query(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.features.list(ListParams())
    assert result.data == []


def test_resolve_forwards_params_and_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/resolve?contactId=cnt_7&featureKey=api_calls",
        json={"data": _RESOLVED},
    )
    with _make_sync() as c:
        resolved = c.features.resolve(ResolveParams(contact_id="cnt_7", feature_key="api_calls"))
    assert resolved.feature.key == "api_calls"
    assert resolved.contributing_subscription_ids == ["sub_1"]
    assert isinstance(resolved.value, ResolvedFeatureQuantity)
    assert resolved.value.quantity == 1000
    assert resolved.value.balance == 850


def test_resolve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/resolve?contactId=cnt_7&featureKey=nope",
        status_code=404,
        json={"error": {"code": "not_found", "message": "unknown feature"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.features.resolve(ResolveParams(contact_id="cnt_7", feature_key="nope"))


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/features/feat_123", json={"data": _SAMPLE})
    with _make_sync() as c:
        feature = c.features.retrieve("feat_123")
    assert feature.key == "api_calls"


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_1?fields=id%2Ckey",
        json={"data": {"id": "feat_1", "key": "api_calls"}},
    )
    with _make_sync() as c:
        feature = c.features.retrieve("feat_1", RetrieveParams(fields="id,key"))
    assert feature.key == "api_calls"


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.features.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.features.retrieve("feat_missing")


def test_create_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features",
        method="POST",
        status_code=201,
        match_json={"key": "api_calls", "name": "API calls", "type": "quantity"},
        json={"data": _SAMPLE},
    )
    with _make_sync() as c:
        feature = c.features.create(CreateBody(key="api_calls", name="API calls", type="quantity"))
    assert feature.id == "feat_123"


def test_create_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.features.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_create_409_conflict(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features",
        method="POST",
        status_code=409,
        json={"error": {"code": "conflict", "message": "feature key exists"}},
    )
    with _make_sync() as c, pytest.raises(ConflictError):
        c.features.create(CreateBody(key="api_calls", name="API calls", type="quantity"))


def test_update_preserves_explicit_null(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123",
        method="PATCH",
        match_json={"name": "API requests", "description": None},
        json={"data": {**_SAMPLE, "name": "API requests", "description": None}},
    )
    with _make_sync() as c:
        feature = c.features.update("feat_123", UpdateBody(name="API requests", description=None))
    assert feature.name == "API requests"


def test_update_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.features.update("", UpdateBody(name="x"))
    assert exc.value.code == "missing_id"


def test_update_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.features.update("feat_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_archive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123/archive",
        method="POST",
        json={"data": {**_SAMPLE, "active": False}},
    )
    with _make_sync() as c:
        feature = c.features.archive("feat_123")
    assert feature.active is False


def test_unarchive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123/unarchive",
        method="POST",
        json={"data": {**_SAMPLE, "active": True}},
    )
    with _make_sync() as c:
        feature = c.features.unarchive("feat_123")
    assert feature.active is True


def test_archive_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.features.archive("")
    assert exc.value.code == "missing_id"


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features?page=0&active=true",
        json={"data": [{"id": "feat_a"}, {"id": "feat_b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/features?page=1&active=true",
        json={"data": [{"id": "feat_c"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [f.id for f in c.features.list_auto_paginate(ListParams(active=True))]
    assert ids == ["feat_a", "feat_b", "feat_c"]


# ────────────────────────────────────────────────────────────────────────────
# Async features
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features", json={"data": [], "hasMore": False}
    )
    async with _make_async() as c:
        r = await c.features.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_resolve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/resolve?contactId=cnt_7&featureKey=api_calls",
        json={"data": _RESOLVED},
    )
    async with _make_async() as c:
        resolved = await c.features.resolve(
            ResolveParams(contact_id="cnt_7", feature_key="api_calls")
        )
    assert resolved.feature.key == "api_calls"


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/features/feat_123", json={"data": _SAMPLE})
    async with _make_async() as c:
        feature = await c.features.retrieve("feat_123")
    assert feature.id == "feat_123"


@pytest.mark.asyncio
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features",
        method="POST",
        status_code=201,
        match_json={
            "key": "plan_tier",
            "name": "Plan tier",
            "type": "enum",
            "enumValues": ["free", "pro"],
        },
        json={"data": _SAMPLE},
    )
    async with _make_async() as c:
        feature = await c.features.create(
            CreateBody(key="plan_tier", name="Plan tier", type="enum", enum_values=["free", "pro"])
        )
    assert feature.id == "feat_123"


@pytest.mark.asyncio
async def test_async_update(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123",
        method="PATCH",
        match_json={"description": None},
        json={"data": {**_SAMPLE, "description": None}},
    )
    async with _make_async() as c:
        feature = await c.features.update("feat_123", UpdateBody(description=None))
    assert feature.id == "feat_123"


@pytest.mark.asyncio
async def test_async_archive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123/archive",
        method="POST",
        json={"data": {**_SAMPLE, "active": False}},
    )
    async with _make_async() as c:
        feature = await c.features.archive("feat_123")
    assert feature.active is False


@pytest.mark.asyncio
async def test_async_unarchive(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features/feat_123/unarchive",
        method="POST",
        json={"data": {**_SAMPLE, "active": True}},
    )
    async with _make_async() as c:
        feature = await c.features.unarchive("feat_123")
    assert feature.active is True


@pytest.mark.asyncio
async def test_async_create_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.features.create(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.features.update("feat_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/features?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/features?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [f.id async for f in c.features.list_auto_paginate()]
    assert ids == ["a", "b", "c"]
