"""Sync and async properties services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.pagination import AsyncIter, Iter
from threecommon.properties.types import (
    CreateBody,
    ListParams,
    ListPropertiesResponse,
    Property,
    UpdateBody,
)

if TYPE_CHECKING:
    from threecommon._core.http_client import AsyncHTTPClient, HTTPClient


def _encode_list_params(params: ListParams | None) -> dict[str, str] | None:
    if params is None:
        return None
    raw = params.model_dump(by_alias=True, exclude_none=True)
    if not raw:
        return None
    return {k: str(v) for k, v in raw.items()}


def _require_id(method: str, property_id: str) -> None:
    if not property_id:
        msg = f"properties.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(property_id: str) -> str:
    return f"/properties/{quote(property_id, safe='')}"


# ----------------------------------------------------------------------------
# Sync
# ----------------------------------------------------------------------------


class PropertiesService:
    """Sync properties service - bound as ``client.properties`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListPropertiesResponse:
        """List the host's properties (one page).

        For full iteration use [list_auto_paginate][PropertiesService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/properties", query=_encode_list_params(params))
        )
        return ListPropertiesResponse.model_validate(body)

    def retrieve(self, property_id: str) -> Property:
        """Retrieve a single property by id."""
        _require_id("retrieve", property_id)
        body = self._http.request(Request(method="GET", path=_path_for(property_id)))
        return Property.model_validate(body["data"])

    def create(self, body: CreateBody) -> Property:
        """Create a new property.

        ``type`` and ``objectType`` can only be set here and cannot be modified
        afterwards. For ``Select One`` and ``Select Multiple`` types, ``options``
        is required and must have at least one entry.
        """
        if body is None:
            raise ValidationError(
                code="missing_body", message="properties.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(Request(method="POST", path="/properties", body=payload))
        return Property.model_validate(response["data"])

    def update(self, property_id: str, body: UpdateBody) -> Property:
        """Apply a partial update to a property. Set ``description`` to ``None`` to clear it."""
        _require_id("update", property_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="properties.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(property_id), body=payload)
        )
        return Property.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Property]:
        """Iterate every property matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Property], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/properties", query=_encode_list_params(page_params))
            )
            response = ListPropertiesResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ----------------------------------------------------------------------------
# Async
# ----------------------------------------------------------------------------


class AsyncPropertiesService:
    """Async properties service - bound as ``client.properties`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListPropertiesResponse:
        body = await self._http.request(
            Request(method="GET", path="/properties", query=_encode_list_params(params))
        )
        return ListPropertiesResponse.model_validate(body)

    async def retrieve(self, property_id: str) -> Property:
        _require_id("retrieve", property_id)
        body = await self._http.request(Request(method="GET", path=_path_for(property_id)))
        return Property.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Property:
        if body is None:
            raise ValidationError(
                code="missing_body", message="properties.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(
            Request(method="POST", path="/properties", body=payload)
        )
        return Property.model_validate(response["data"])

    async def update(self, property_id: str, body: UpdateBody) -> Property:
        _require_id("update", property_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="properties.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(property_id), body=payload)
        )
        return Property.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Property]:
        """Async iterate every property matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Property], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/properties", query=_encode_list_params(page_params))
            )
            response = ListPropertiesResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
