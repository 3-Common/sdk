"""Sync and async features services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.features.types import (
    CreateBody,
    Feature,
    ListFeaturesResponse,
    ListParams,
    ResolvedFeature,
    ResolveParams,
    RetrieveParams,
    UpdateBody,
)
from threecommon.pagination import AsyncIter, Iter

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


def _encode_resolve_params(params: ResolveParams) -> dict[str, str]:
    raw = params.model_dump(by_alias=True, exclude_none=True)
    return {k: str(v) for k, v in raw.items()}


def _require_id(method: str, feature_id: str) -> None:
    if not feature_id:
        msg = f"features.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(feature_id: str) -> str:
    return f"/features/{quote(feature_id, safe='')}"


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class FeaturesService:
    """Sync features service — bound as ``client.features`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListFeaturesResponse:
        """List the host's feature catalog (one page).

        For full iteration use [list_auto_paginate][FeaturesService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/features", query=_encode_list_params(params))
        )
        return ListFeaturesResponse.model_validate(body)

    def resolve(self, params: ResolveParams) -> ResolvedFeature:
        """Resolve a feature's live value for a customer.

        Walks active subscriptions -> prices -> feature grants. Raises
        [NotFoundError][threecommon.NotFoundError] if the feature key is unknown.
        """
        body = self._http.request(
            Request(method="GET", path="/features/resolve", query=_encode_resolve_params(params))
        )
        return ResolvedFeature.model_validate(body["data"])

    def retrieve(self, feature_id: str, params: RetrieveParams | None = None) -> Feature:
        """Retrieve a single feature by id."""
        _require_id("retrieve", feature_id)
        body = self._http.request(
            Request(method="GET", path=_path_for(feature_id), query=_encode_retrieve_params(params))
        )
        return Feature.model_validate(body["data"])

    def create(self, body: CreateBody) -> Feature:
        """Create a feature in the catalog."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="features.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(Request(method="POST", path="/features", body=payload))
        return Feature.model_validate(response["data"])

    def update(self, feature_id: str, body: UpdateBody) -> Feature:
        """Apply a partial update to a feature. Set a field to ``None`` to clear it."""
        _require_id("update", feature_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="features.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(feature_id), body=payload)
        )
        return Feature.model_validate(response["data"])

    def archive(self, feature_id: str) -> Feature:
        """Soft-archive a feature. Idempotent."""
        _require_id("archive", feature_id)
        response = self._http.request(
            Request(method="POST", path=f"{_path_for(feature_id)}/archive")
        )
        return Feature.model_validate(response["data"])

    def unarchive(self, feature_id: str) -> Feature:
        """Reactivate a previously archived feature. Idempotent."""
        _require_id("unarchive", feature_id)
        response = self._http.request(
            Request(method="POST", path=f"{_path_for(feature_id)}/unarchive")
        )
        return Feature.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Feature]:
        """Iterate every feature matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Feature], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path="/features", query=_encode_list_params(page_params))
            )
            response = ListFeaturesResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncFeaturesService:
    """Async features service — bound as ``client.features`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListFeaturesResponse:
        body = await self._http.request(
            Request(method="GET", path="/features", query=_encode_list_params(params))
        )
        return ListFeaturesResponse.model_validate(body)

    async def resolve(self, params: ResolveParams) -> ResolvedFeature:
        body = await self._http.request(
            Request(method="GET", path="/features/resolve", query=_encode_resolve_params(params))
        )
        return ResolvedFeature.model_validate(body["data"])

    async def retrieve(self, feature_id: str, params: RetrieveParams | None = None) -> Feature:
        _require_id("retrieve", feature_id)
        body = await self._http.request(
            Request(method="GET", path=_path_for(feature_id), query=_encode_retrieve_params(params))
        )
        return Feature.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Feature:
        if body is None:
            raise ValidationError(
                code="missing_body", message="features.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(Request(method="POST", path="/features", body=payload))
        return Feature.model_validate(response["data"])

    async def update(self, feature_id: str, body: UpdateBody) -> Feature:
        _require_id("update", feature_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="features.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_unset=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(feature_id), body=payload)
        )
        return Feature.model_validate(response["data"])

    async def archive(self, feature_id: str) -> Feature:
        _require_id("archive", feature_id)
        response = await self._http.request(
            Request(method="POST", path=f"{_path_for(feature_id)}/archive")
        )
        return Feature.model_validate(response["data"])

    async def unarchive(self, feature_id: str) -> Feature:
        _require_id("unarchive", feature_id)
        response = await self._http.request(
            Request(method="POST", path=f"{_path_for(feature_id)}/unarchive")
        )
        return Feature.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Feature]:
        """Async iterate every feature matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Feature], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path="/features", query=_encode_list_params(page_params))
            )
            response = ListFeaturesResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
