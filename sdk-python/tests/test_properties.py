"""Properties service tests - sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.properties import (
    CreateBody,
    ListParams,
    Property,
    PropertyOption,
    UpdateBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


_TEXT = {
    "type": "Text",
    "id": "prop_1",
    "name": "Dietary notes",
    "status": "active",
    "objectType": "contact",
}

_SELECT_ONE = {
    "type": "Select One",
    "id": "prop_2",
    "name": "T-shirt size",
    "status": "active",
    "objectType": "contact",
    "options": [
        {"value": "s", "label": "Small"},
        {"value": "m", "label": "Medium"},
    ],
}


# ----------------------------------------------------------------------------
# Sync properties
# ----------------------------------------------------------------------------


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties?objectType=contact&status=active&pageSize=10",
        json={"data": [_TEXT, _SELECT_ONE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.properties.list(ListParams(object_type="contact", status="active", page_size=10))
    assert len(result.data) == 2
    assert result.data[0].id == "prop_1"
    assert result.data[0].object_type == "contact"
    assert result.data[1].options is not None
    assert result.data[1].options[0].value == "s"
    assert result.has_more is False


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.properties.list()
    assert result.data == []


def test_list_empty_params_omits_query(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.properties.list(ListParams())
    assert result.data == []


def test_list_forwards_filters(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(json={"data": [], "hasMore": False})
    with _make_sync() as c:
        c.properties.list(
            ListParams(property_type="Select One", sort="name", order="desc", search="size")
        )
    req = httpx_mock.get_requests()[0]
    assert req.url.params.get("propertyType") == "Select One"
    assert req.url.params.get("sort") == "name"
    assert req.url.params.get("order") == "desc"
    assert req.url.params.get("search") == "size"


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties/prop_2", json={"data": _SELECT_ONE}
    )
    with _make_sync() as c:
        prop = c.properties.retrieve("prop_2")
    assert prop.id == "prop_2"
    assert prop.type == "Select One"
    assert prop.options is not None
    assert prop.options[1].label == "Medium"


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.properties.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties/prop_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.properties.retrieve("prop_missing")


def test_property_text_has_no_options() -> None:
    prop = Property.model_validate(_TEXT)
    assert prop.options is None
    assert prop.description is None


def test_create_sends_camelcase_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties",
        method="POST",
        match_json={
            "type": "Select One",
            "name": "T-shirt size",
            "status": "active",
            "objectType": "contact",
            "options": [
                {"value": "s", "label": "Small"},
                {"value": "m", "label": "Medium"},
            ],
        },
        json={"data": _SELECT_ONE},
    )
    with _make_sync() as c:
        prop = c.properties.create(
            CreateBody(
                type="Select One",
                name="T-shirt size",
                status="active",
                object_type="contact",
                options=[
                    PropertyOption(value="s", label="Small"),
                    PropertyOption(value="m", label="Medium"),
                ],
            )
        )
    assert prop.id == "prop_2"


def test_create_omits_unset_options(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties",
        method="POST",
        match_json={
            "type": "Text",
            "name": "Dietary notes",
            "status": "active",
            "objectType": "contact",
        },
        json={"data": _TEXT},
    )
    with _make_sync() as c:
        c.properties.create(
            CreateBody(type="Text", name="Dietary notes", status="active", object_type="contact")
        )


def test_create_422_validation(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties",
        method="POST",
        status_code=422,
        json={"error": {"code": "validation_error", "message": "options required"}},
    )
    with _make_sync() as c, pytest.raises(ValidationError):
        c.properties.create(
            CreateBody(
                type="Select One", name="T-shirt size", status="active", object_type="contact"
            )
        )


def test_create_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.properties.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_update_preserves_explicit_null(httpx_mock: HTTPXMock) -> None:
    # description=None clears the field; unset fields (status, options) are omitted.
    httpx_mock.add_response(
        url="http://test.local/v1/properties/prop_1",
        method="PATCH",
        match_json={"name": "Allergies", "description": None},
        json={"data": {**_TEXT, "name": "Allergies"}},
    )
    with _make_sync() as c:
        prop = c.properties.update("prop_1", UpdateBody(name="Allergies", description=None))
    assert prop.name == "Allergies"


def test_update_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.properties.update("", UpdateBody(name="x"))
    assert exc.value.code == "missing_id"


def test_update_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.properties.update("prop_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties?page=0&objectType=contact",
        json={
            "data": [{**_TEXT, "id": "prop_a"}, {**_TEXT, "id": "prop_b"}],
            "hasMore": True,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/properties?page=1&objectType=contact",
        json={"data": [{**_TEXT, "id": "prop_c"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [p.id for p in c.properties.list_auto_paginate(ListParams(object_type="contact"))]
    assert ids == ["prop_a", "prop_b", "prop_c"]


def test_list_auto_paginate_no_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties?page=0",
        json={"data": [{**_TEXT, "id": "only"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [p.id for p in c.properties.list_auto_paginate()]
    assert ids == ["only"]


# ----------------------------------------------------------------------------
# Async properties
# ----------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties", json={"data": [], "hasMore": False}
    )
    async with _make_async() as c:
        r = await c.properties.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/properties/prop_1", json={"data": _TEXT})
    async with _make_async() as c:
        prop = await c.properties.retrieve("prop_1")
    assert prop.id == "prop_1"
    assert prop.name == "Dietary notes"


@pytest.mark.asyncio
async def test_async_retrieve_requires_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.properties.retrieve("")
        assert exc.value.code == "missing_id"


@pytest.mark.asyncio
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties",
        method="POST",
        json={"data": _TEXT},
    )
    async with _make_async() as c:
        prop = await c.properties.create(
            CreateBody(type="Text", name="Dietary notes", status="active", object_type="contact")
        )
    assert prop.id == "prop_1"


@pytest.mark.asyncio
async def test_async_create_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.properties.create(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_update(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties/prop_1",
        method="PATCH",
        match_json={"status": "archived"},
        json={"data": {**_TEXT, "status": "archived"}},
    )
    async with _make_async() as c:
        prop = await c.properties.update("prop_1", UpdateBody(status="archived"))
    assert prop.status == "archived"


@pytest.mark.asyncio
async def test_async_update_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.properties.update("prop_1", None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/properties?page=0",
        json={"data": [{**_TEXT, "id": "a"}, {**_TEXT, "id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/properties?page=1",
        json={"data": [{**_TEXT, "id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [p.id async for p in c.properties.list_auto_paginate()]
    assert ids == ["a", "b", "c"]
