"""Sync and async entitlements services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.entitlements.types import (
    ConsumeBody,
    Entitlement,
    GrantBody,
    ListEntitlementsResponse,
    ListParams,
    LookupParams,
    RetrieveParams,
)
from threecommon.errors.classes import ValidationError
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


def _encode_lookup_params(params: LookupParams) -> dict[str, str]:
    raw = params.model_dump(by_alias=True, exclude_none=True)
    return {k: str(v) for k, v in raw.items()}


def _require_id(method: str, entitlement_id: str) -> None:
    if not entitlement_id:
        msg = f"entitlements.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _require_lookup_params(params: LookupParams) -> None:
    if not params.contact_id:
        raise ValidationError(
            code="missing_contact_id",
            message="entitlements.lookup: contact_id must be a non-empty string",
        )
    if not params.feature_key:
        raise ValidationError(
            code="missing_feature_key",
            message="entitlements.lookup: feature_key must be a non-empty string",
        )


def _path_for(entitlement_id: str) -> str:
    return f"/entitlements/{quote(entitlement_id, safe='')}"


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class EntitlementsService:
    """Sync entitlements service — bound as ``client.entitlements`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListEntitlementsResponse:
        """List the host's entitlement balance records (one page).

        For full iteration use
        [list_auto_paginate][EntitlementsService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/entitlements", query=_encode_list_params(params))
        )
        return ListEntitlementsResponse.model_validate(body)

    def retrieve(self, entitlement_id: str, params: RetrieveParams | None = None) -> Entitlement:
        """Retrieve a single entitlement record by id, including grant history."""
        _require_id("retrieve", entitlement_id)
        body = self._http.request(
            Request(
                method="GET",
                path=_path_for(entitlement_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Entitlement.model_validate(body["data"])

    def lookup(self, params: LookupParams) -> Entitlement:
        """Look up the unique entitlement for a ``(contact_id, feature_key)`` pair.

        Raises [NotFoundError][threecommon.NotFoundError] if no record exists yet.
        """
        _require_lookup_params(params)
        body = self._http.request(
            Request(method="GET", path="/entitlements/lookup", query=_encode_lookup_params(params))
        )
        return Entitlement.model_validate(body["data"])

    def grant(self, body: GrantBody) -> Entitlement:
        """Add a manual entitlement grant. Idempotent on ``grant_id``."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="entitlements.grant: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path="/entitlements/grants", body=payload)
        )
        return Entitlement.model_validate(response["data"])

    def consume(self, body: ConsumeBody) -> Entitlement:
        """Debit units from a customer's entitlement balance.

        Raises [ConflictError][threecommon.ConflictError] on insufficient balance.
        """
        if body is None:
            raise ValidationError(
                code="missing_body", message="entitlements.consume: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path="/entitlements/consume", body=payload)
        )
        return Entitlement.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Entitlement]:
        """Iterate every entitlement matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Entitlement], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/entitlements", query=_encode_list_params(page_params))
            )
            response = ListEntitlementsResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncEntitlementsService:
    """Async entitlements service — bound as ``client.entitlements`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListEntitlementsResponse:
        body = await self._http.request(
            Request(method="GET", path="/entitlements", query=_encode_list_params(params))
        )
        return ListEntitlementsResponse.model_validate(body)

    async def retrieve(
        self, entitlement_id: str, params: RetrieveParams | None = None
    ) -> Entitlement:
        _require_id("retrieve", entitlement_id)
        body = await self._http.request(
            Request(
                method="GET",
                path=_path_for(entitlement_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Entitlement.model_validate(body["data"])

    async def lookup(self, params: LookupParams) -> Entitlement:
        _require_lookup_params(params)
        body = await self._http.request(
            Request(method="GET", path="/entitlements/lookup", query=_encode_lookup_params(params))
        )
        return Entitlement.model_validate(body["data"])

    async def grant(self, body: GrantBody) -> Entitlement:
        if body is None:
            raise ValidationError(
                code="missing_body", message="entitlements.grant: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path="/entitlements/grants", body=payload)
        )
        return Entitlement.model_validate(response["data"])

    async def consume(self, body: ConsumeBody) -> Entitlement:
        if body is None:
            raise ValidationError(
                code="missing_body", message="entitlements.consume: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path="/entitlements/consume", body=payload)
        )
        return Entitlement.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Entitlement]:
        """Async iterate every entitlement matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Entitlement], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/entitlements", query=_encode_list_params(page_params))
            )
            response = ListEntitlementsResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
