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
    AttachPaymentMethodBody,
    AttachPaymentMethodResult,
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
    PaymentMethod,
    PaymentMethodSetupIntent,
    RemovedPaymentMethod,
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


def _require_method_id(method: str, method_id: str) -> None:
    if not method_id:
        msg = f"contacts.{method}: method_id must be a non-empty string"
        raise ValidationError(code="missing_method_id", message=msg)


def _path_for(contact_id: str) -> str:
    return f"/contacts/{quote(contact_id, safe='')}"


def _activity_path(contact_id: str) -> str:
    return f"{_path_for(contact_id)}/activity"


def _payment_methods_path(contact_id: str) -> str:
    return f"{_path_for(contact_id)}/payment-methods"


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

    def retrieve_payment_method(self, contact_id: str) -> PaymentMethod | None:
        """Retrieve the saved card on file for a contact.

        Returns ``None`` when the contact has no card saved. One card is
        supported per contact.
        """
        _require_id("retrieve_payment_method", contact_id)
        body = self._http.request(Request(method="GET", path=_payment_methods_path(contact_id)))
        data = body["data"]
        if data is None:
            return None
        return PaymentMethod.model_validate(data)

    def attach_payment_method(
        self, contact_id: str, body: AttachPaymentMethodBody
    ) -> AttachPaymentMethodResult:
        """Persist the card from a confirmed SetupIntent against the contact.

        Replaces any existing card; ``replaced_existing`` reports whether one
        was overwritten. The SetupIntent is re-verified server-side.
        """
        _require_id("attach_payment_method", contact_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="contacts.attach_payment_method: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=_payment_methods_path(contact_id), body=payload)
        )
        return AttachPaymentMethodResult.model_validate(response)

    def create_payment_method_setup_intent(self, contact_id: str) -> PaymentMethodSetupIntent:
        """Begin saving a card for a contact.

        Returns a Stripe SetupIntent ``client_secret`` to confirm client-side;
        afterwards call ``attach_payment_method`` with the ``setup_intent_id``.
        """
        _require_id("create_payment_method_setup_intent", contact_id)
        response = self._http.request(
            Request(method="POST", path=f"{_payment_methods_path(contact_id)}/setup-intent")
        )
        return PaymentMethodSetupIntent.model_validate(response["data"])

    def remove_payment_method(self, contact_id: str, method_id: str) -> RemovedPaymentMethod:
        """Detach the saved card from Stripe and remove it from the contact."""
        _require_id("remove_payment_method", contact_id)
        _require_method_id("remove_payment_method", method_id)
        response = self._http.request(
            Request(
                method="DELETE",
                path=f"{_payment_methods_path(contact_id)}/{quote(method_id, safe='')}",
            )
        )
        return RemovedPaymentMethod.model_validate(response["data"])

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

    async def retrieve_payment_method(self, contact_id: str) -> PaymentMethod | None:
        _require_id("retrieve_payment_method", contact_id)
        body = await self._http.request(
            Request(method="GET", path=_payment_methods_path(contact_id))
        )
        data = body["data"]
        if data is None:
            return None
        return PaymentMethod.model_validate(data)

    async def attach_payment_method(
        self, contact_id: str, body: AttachPaymentMethodBody
    ) -> AttachPaymentMethodResult:
        _require_id("attach_payment_method", contact_id)
        if body is None:
            raise ValidationError(
                code="missing_body",
                message="contacts.attach_payment_method: body must be non-None",
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_payment_methods_path(contact_id), body=payload)
        )
        return AttachPaymentMethodResult.model_validate(response)

    async def create_payment_method_setup_intent(self, contact_id: str) -> PaymentMethodSetupIntent:
        _require_id("create_payment_method_setup_intent", contact_id)
        response = await self._http.request(
            Request(method="POST", path=f"{_payment_methods_path(contact_id)}/setup-intent")
        )
        return PaymentMethodSetupIntent.model_validate(response["data"])

    async def remove_payment_method(self, contact_id: str, method_id: str) -> RemovedPaymentMethod:
        _require_id("remove_payment_method", contact_id)
        _require_method_id("remove_payment_method", method_id)
        response = await self._http.request(
            Request(
                method="DELETE",
                path=f"{_payment_methods_path(contact_id)}/{quote(method_id, safe='')}",
            )
        )
        return RemovedPaymentMethod.model_validate(response["data"])

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
