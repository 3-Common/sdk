"""Sync and async events services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.events.types import (
    Event,
    ListEventsResponse,
    ListParams,
    RetrieveParams,
    UpdateBody,
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


def _require_id(method: str, event_id: str) -> None:
    if not event_id:
        msg = f"events.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(event_id: str) -> str:
    return f"/events/{quote(event_id, safe='')}"


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class EventsService:
    """Sync events service — bound as ``client.events`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListEventsResponse:
        """List the authenticated host's events (one page).

        For full iteration use [list_auto_paginate][EventsService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/events", query=_encode_list_params(params))
        )
        return ListEventsResponse.model_validate(body)

    def retrieve(self, event_id: str, params: RetrieveParams | None = None) -> Event:
        """Retrieve a single event by ID."""
        _require_id("retrieve", event_id)
        body = self._http.request(
            Request(method="GET", path=_path_for(event_id), query=_encode_retrieve_params(params))
        )
        return Event.model_validate(body["data"])

    def update(self, event_id: str, body: UpdateBody) -> Event:
        """Update an event's basic fields. Only fields you provide change."""
        _require_id("update", event_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="events.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(event_id), body=payload)
        )
        return Event.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Event]:
        """Iterate every event matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Event], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/events", query=_encode_list_params(page_params))
            )
            response = ListEventsResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncEventsService:
    """Async events service — bound as ``client.events`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListEventsResponse:
        body = await self._http.request(
            Request(method="GET", path="/events", query=_encode_list_params(params))
        )
        return ListEventsResponse.model_validate(body)

    async def retrieve(self, event_id: str, params: RetrieveParams | None = None) -> Event:
        _require_id("retrieve", event_id)
        body = await self._http.request(
            Request(method="GET", path=_path_for(event_id), query=_encode_retrieve_params(params))
        )
        return Event.model_validate(body["data"])

    async def update(self, event_id: str, body: UpdateBody) -> Event:
        _require_id("update", event_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="events.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(event_id), body=payload)
        )
        return Event.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Event]:
        """Async iterate every event matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Event], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/events", query=_encode_list_params(page_params))
            )
            response = ListEventsResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
