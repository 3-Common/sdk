"""Sync and async contacts services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.contacts.types import (
    ActivityListParams,
    BulkUpsertBody,
    BulkUpsertResult,
    Contact,
    ContactActivity,
    ContactWithOrderDetails,
    CountResult,
    CreateBody,
    DeleteResult,
    ListActivityResponse,
    ListContactsResponse,
    ListParams,
    UpdateBody,
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


def _encode_activity_params(params: ActivityListParams | None) -> dict[str, str] | None:
    if params is None:
        return None
    raw = params.model_dump(by_alias=True, exclude_none=True)
    if not raw:
        return None
    return {k: str(v) for k, v in raw.items()}


def _require_id(method: str, contact_id: str) -> None:
    if not contact_id:
        msg = f"contacts.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(contact_id: str) -> str:
    return f"/contacts/{quote(contact_id, safe='')}"


def _activity_path(contact_id: str) -> str:
    return f"{_path_for(contact_id)}/activity"


# ---------------
# Sync
# ---------------


class ContactsService:
    """Sync contacts service — bound as ``client.contacts`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListContactsResponse:
        """List the host's contacts (one page).

        For full iteration use [list_auto_paginate][ContactsService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/contacts", query=_encode_list_params(params))
        )
        return ListContactsResponse.model_validate(body)

    def count(self) -> CountResult:
        """Return the total contact count for the host."""
        body = self._http.request(Request(method="GET", path="/contacts/count"))
        return CountResult.model_validate(body["data"])

    def retrieve(self, contact_id: str) -> Contact:
        """Retrieve a single contact by id."""
        _require_id("retrieve", contact_id)
        body = self._http.request(Request(method="GET", path=_path_for(contact_id)))
        return Contact.model_validate(body["data"])

    def create(self, body: CreateBody) -> Contact:
        """Create a new contact. Raises ``ConflictError`` on duplicate email."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(Request(method="POST", path="/contacts", body=payload))
        return Contact.model_validate(response["data"])

    def update(self, contact_id: str, body: UpdateBody) -> ContactWithOrderDetails:
        """Update a contact. Returns the richer order-details projection."""
        _require_id("update", contact_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(contact_id), body=payload)
        )
        return ContactWithOrderDetails.model_validate(response["data"])

    def delete(self, contact_id: str) -> DeleteResult:
        """Permanently remove a contact. Echoes the removed contact's id."""
        _require_id("delete", contact_id)
        response = self._http.request(Request(method="DELETE", path=_path_for(contact_id)))
        return DeleteResult.model_validate(response["data"])

    def bulk_upsert(self, body: BulkUpsertBody) -> BulkUpsertResult:
        """Bulk-upsert up to 10,000 contacts in one round-trip.

        Deduplicated server-side by email; existing rows are updated rather
        than rejected.
        """
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.bulk_upsert: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(Request(method="POST", path="/contacts/bulk", body=payload))
        return BulkUpsertResult.model_validate(response["data"])

    def list_activity(
        self, contact_id: str, params: ActivityListParams | None = None
    ) -> ListActivityResponse:
        """Paginated activity log for a contact."""
        _require_id("list_activity", contact_id)
        body = self._http.request(
            Request(
                method="GET",
                path=_activity_path(contact_id),
                query=_encode_activity_params(params),
            )
        )
        return ListActivityResponse.model_validate(body)

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Contact]:
        """Iterate every contact matching ``params``, paging automatically."""
        start_page = (
            params.page_number if params is not None and params.page_number is not None else 0
        )

        def fetch(page: int) -> tuple[list[Contact], bool]:
            page_params = (
                params.model_copy(update={"page_number": page})
                if params is not None
                else ListParams(page_number=page)
            )
            body = self._http.request(
                Request(method="GET", path="/contacts", query=_encode_list_params(page_params))
            )
            response = ListContactsResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)

    def list_activity_auto_paginate(
        self, contact_id: str, params: ActivityListParams | None = None
    ) -> Iter[ContactActivity]:
        """Iterate every activity record for a contact, paging automatically."""
        _require_id("list_activity_auto_paginate", contact_id)
        start_page = (
            params.page_number if params is not None and params.page_number is not None else 0
        )

        def fetch(page: int) -> tuple[list[ContactActivity], bool]:
            page_params = (
                params.model_copy(update={"page_number": page})
                if params is not None
                else ActivityListParams(page_number=page)
            )
            body = self._http.request(
                Request(
                    method="GET",
                    path=_activity_path(contact_id),
                    query=_encode_activity_params(page_params),
                )
            )
            response = ListActivityResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ---------------
# Async
# ---------------


class AsyncContactsService:
    """Async contacts service — bound as ``client.contacts`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListContactsResponse:
        body = await self._http.request(
            Request(method="GET", path="/contacts", query=_encode_list_params(params))
        )
        return ListContactsResponse.model_validate(body)

    async def count(self) -> CountResult:
        body = await self._http.request(Request(method="GET", path="/contacts/count"))
        return CountResult.model_validate(body["data"])

    async def retrieve(self, contact_id: str) -> Contact:
        _require_id("retrieve", contact_id)
        body = await self._http.request(Request(method="GET", path=_path_for(contact_id)))
        return Contact.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Contact:
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(Request(method="POST", path="/contacts", body=payload))
        return Contact.model_validate(response["data"])

    async def update(self, contact_id: str, body: UpdateBody) -> ContactWithOrderDetails:
        _require_id("update", contact_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(contact_id), body=payload)
        )
        return ContactWithOrderDetails.model_validate(response["data"])

    async def delete(self, contact_id: str) -> DeleteResult:
        _require_id("delete", contact_id)
        response = await self._http.request(Request(method="DELETE", path=_path_for(contact_id)))
        return DeleteResult.model_validate(response["data"])

    async def bulk_upsert(self, body: BulkUpsertBody) -> BulkUpsertResult:
        if body is None:
            raise ValidationError(
                code="missing_body", message="contacts.bulk_upsert: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path="/contacts/bulk", body=payload)
        )
        return BulkUpsertResult.model_validate(response["data"])

    async def list_activity(
        self, contact_id: str, params: ActivityListParams | None = None
    ) -> ListActivityResponse:
        _require_id("list_activity", contact_id)
        body = await self._http.request(
            Request(
                method="GET",
                path=_activity_path(contact_id),
                query=_encode_activity_params(params),
            )
        )
        return ListActivityResponse.model_validate(body)

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Contact]:
        """Async iterate every contact matching ``params``."""
        start_page = (
            params.page_number if params is not None and params.page_number is not None else 0
        )
        http = self._http

        async def fetch(page: int) -> tuple[list[Contact], bool]:
            page_params = (
                params.model_copy(update={"page_number": page})
                if params is not None
                else ListParams(page_number=page)
            )
            body = await http.request(
                Request(method="GET", path="/contacts", query=_encode_list_params(page_params))
            )
            response = ListContactsResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)

    def list_activity_auto_paginate(
        self, contact_id: str, params: ActivityListParams | None = None
    ) -> AsyncIter[ContactActivity]:
        """Async iterate every activity record for a contact."""
        _require_id("list_activity_auto_paginate", contact_id)
        start_page = (
            params.page_number if params is not None and params.page_number is not None else 0
        )
        http = self._http

        async def fetch(page: int) -> tuple[list[ContactActivity], bool]:
            page_params = (
                params.model_copy(update={"page_number": page})
                if params is not None
                else ActivityListParams(page_number=page)
            )
            body = await http.request(
                Request(
                    method="GET",
                    path=_activity_path(contact_id),
                    query=_encode_activity_params(page_params),
                )
            )
            response = ListActivityResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
