"""Sync and async prices services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.pagination import AsyncIter, Iter
from threecommon.prices.types import (
    CreateBody,
    ListParams,
    ListPricesResponse,
    Price,
    RetrieveParams,
    UpdateBody,
)

if TYPE_CHECKING:
    from threecommon._core.http_client import AsyncHTTPClient, HTTPClient


def _qval(value: object) -> str:
    # The wire (and every other SDK) renders query booleans lowercase.
    if isinstance(value, bool):
        return "true" if value else "false"
    return str(value)


def _encode_list_params(params: ListParams | None) -> dict[str, str] | None:
    if params is None:
        return None
    raw = params.model_dump(by_alias=True, exclude_none=True)
    if not raw:
        return None
    return {k: _qval(v) for k, v in raw.items()}


def _encode_retrieve_params(params: RetrieveParams | None) -> dict[str, str] | None:
    if params is None or params.fields is None:
        return None
    return {"fields": params.fields}


def _require_id(method: str, price_id: str) -> None:
    if not price_id:
        msg = f"prices.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(price_id: str) -> str:
    return f"/prices/{quote(price_id, safe='')}"


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class PricesService:
    """Sync prices service — bound as ``client.prices`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListPricesResponse:
        """List the host's prices (one page).

        For full iteration use [list_auto_paginate][PricesService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/prices", query=_encode_list_params(params))
        )
        return ListPricesResponse.model_validate(body)

    def retrieve(self, price_id: str, params: RetrieveParams | None = None) -> Price:
        """Retrieve a single price by id."""
        _require_id("retrieve", price_id)
        body = self._http.request(
            Request(method="GET", path=_path_for(price_id), query=_encode_retrieve_params(params))
        )
        return Price.model_validate(body["data"])

    def create(self, body: CreateBody) -> Price:
        """Create a price for a product."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="prices.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(Request(method="POST", path="/prices", body=payload))
        return Price.model_validate(response["data"])

    def update(self, price_id: str, body: UpdateBody) -> Price:
        """Apply a partial update to a price. Set a field to ``None`` to clear it."""
        _require_id("update", price_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="prices.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(price_id), body=payload)
        )
        return Price.model_validate(response["data"])

    def archive(self, price_id: str) -> Price:
        """Soft-archive a price. Idempotent."""
        _require_id("archive", price_id)
        response = self._http.request(Request(method="POST", path=f"{_path_for(price_id)}/archive"))
        return Price.model_validate(response["data"])

    def unarchive(self, price_id: str) -> Price:
        """Reactivate a previously archived price. Idempotent."""
        _require_id("unarchive", price_id)
        response = self._http.request(
            Request(method="POST", path=f"{_path_for(price_id)}/unarchive")
        )
        return Price.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Price]:
        """Iterate every price matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Price], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/prices", query=_encode_list_params(page_params))
            )
            response = ListPricesResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncPricesService:
    """Async prices service — bound as ``client.prices`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListPricesResponse:
        body = await self._http.request(
            Request(method="GET", path="/prices", query=_encode_list_params(params))
        )
        return ListPricesResponse.model_validate(body)

    async def retrieve(self, price_id: str, params: RetrieveParams | None = None) -> Price:
        _require_id("retrieve", price_id)
        body = await self._http.request(
            Request(method="GET", path=_path_for(price_id), query=_encode_retrieve_params(params))
        )
        return Price.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Price:
        if body is None:
            raise ValidationError(
                code="missing_body", message="prices.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(Request(method="POST", path="/prices", body=payload))
        return Price.model_validate(response["data"])

    async def update(self, price_id: str, body: UpdateBody) -> Price:
        _require_id("update", price_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="prices.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(price_id), body=payload)
        )
        return Price.model_validate(response["data"])

    async def archive(self, price_id: str) -> Price:
        _require_id("archive", price_id)
        response = await self._http.request(
            Request(method="POST", path=f"{_path_for(price_id)}/archive")
        )
        return Price.model_validate(response["data"])

    async def unarchive(self, price_id: str) -> Price:
        _require_id("unarchive", price_id)
        response = await self._http.request(
            Request(method="POST", path=f"{_path_for(price_id)}/unarchive")
        )
        return Price.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Price]:
        """Async iterate every price matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Price], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/prices", query=_encode_list_params(page_params))
            )
            response = ListPricesResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
