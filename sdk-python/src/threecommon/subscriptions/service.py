"""Sync and async subscriptions services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.pagination import AsyncIter, Iter
from threecommon.subscriptions.types import (
    BillSubscriptionResult,
    CancelBody,
    CancelImmediatelyBody,
    CreateBody,
    ListParams,
    ListSubscriptionsResponse,
    RenewSubscriptionResult,
    RetrieveParams,
    Subscription,
    SubscriptionInvoicePreview,
    UpdateBody,
    UpdateSubscriptionResult,
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


def _encode_retrieve_params(params: RetrieveParams | None) -> dict[str, str] | None:
    if params is None or params.fields is None:
        return None
    return {"fields": params.fields}


def _require_id(method: str, subscription_id: str) -> None:
    if not subscription_id:
        msg = f"subscriptions.{method}: id must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _path_for(subscription_id: str) -> str:
    return f"/subscriptions/{quote(subscription_id, safe='')}"


def _action_path(subscription_id: str, action: str) -> str:
    return f"{_path_for(subscription_id)}/{action}"


def _build_update_result(response: dict[str, Any]) -> UpdateSubscriptionResult:
    payload: dict[str, Any] = {
        "subscription": response["data"],
        "proration": response["proration"],
    }
    if response.get("invoice") is not None:
        payload["invoice"] = response["invoice"]
    return UpdateSubscriptionResult.model_validate(payload)


def _build_bill_result(response: dict[str, Any]) -> BillSubscriptionResult:
    return BillSubscriptionResult.model_validate(
        {"subscription": response["data"], "invoice": response["invoice"]}
    )


def _build_renew_result(response: dict[str, Any]) -> RenewSubscriptionResult:
    payload: dict[str, Any] = {"subscription": response["data"]}
    if response.get("invoice") is not None:
        payload["invoice"] = response["invoice"]
    return RenewSubscriptionResult.model_validate(payload)


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


class SubscriptionsService:
    """Sync subscriptions service — bound as ``client.subscriptions`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListSubscriptionsResponse:
        """List the host's subscriptions (one page).

        For full iteration use
        [list_auto_paginate][SubscriptionsService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path="/subscriptions", query=_encode_list_params(params))
        )
        return ListSubscriptionsResponse.model_validate(body)

    def retrieve(self, subscription_id: str, params: RetrieveParams | None = None) -> Subscription:
        """Retrieve a single subscription by id."""
        _require_id("retrieve", subscription_id)
        body = self._http.request(
            Request(
                method="GET",
                path=_path_for(subscription_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Subscription.model_validate(body["data"])

    def create(self, body: CreateBody) -> Subscription:
        """Create a new subscription against an active recurring Price."""
        if body is None:
            raise ValidationError(
                code="missing_body", message="subscriptions.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(Request(method="POST", path="/subscriptions", body=payload))
        return Subscription.model_validate(response["data"])

    def update(self, subscription_id: str, body: UpdateBody) -> UpdateSubscriptionResult:
        """Apply a mid-cycle price/quantity change with Stripe-style daily proration."""
        _require_id("update", subscription_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="subscriptions.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_path_for(subscription_id), body=payload)
        )
        return _build_update_result(response)

    def activate(self, subscription_id: str) -> Subscription:
        """Transition an incomplete or trialing subscription to ``active``."""
        _require_id("activate", subscription_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "activate"), body={})
        )
        return Subscription.model_validate(response["data"])

    def cancel(self, subscription_id: str, body: CancelBody | None = None) -> Subscription:
        """Schedule cancellation at the end of the current period. Idempotent."""
        _require_id("cancel", subscription_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "cancel"), body=payload)
        )
        return Subscription.model_validate(response["data"])

    def cancel_immediately(
        self, subscription_id: str, body: CancelImmediatelyBody | None = None
    ) -> Subscription:
        """Admin override — terminate the subscription immediately."""
        _require_id("cancel_immediately", subscription_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = self._http.request(
            Request(
                method="POST",
                path=_action_path(subscription_id, "cancel-immediately"),
                body=payload,
            )
        )
        return Subscription.model_validate(response["data"])

    def mark_unpaid(self, subscription_id: str) -> Subscription:
        """Admin override — mark a subscription ``unpaid`` (terminal)."""
        _require_id("mark_unpaid", subscription_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "mark-unpaid"), body={})
        )
        return Subscription.model_validate(response["data"])

    def bill(self, subscription_id: str) -> BillSubscriptionResult:
        """Generate a draft invoice for the current period without advancing it."""
        _require_id("bill", subscription_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "bill"), body={})
        )
        return _build_bill_result(response)

    def renew(self, subscription_id: str) -> RenewSubscriptionResult:
        """Advance the subscription to its next billing period and generate an invoice."""
        _require_id("renew", subscription_id)
        response = self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "renew"), body={})
        )
        return _build_renew_result(response)

    def preview_upcoming_invoice(self, subscription_id: str) -> SubscriptionInvoicePreview | None:
        """Preview the invoice the next renewal will generate.

        Returns ``None`` when the subscription is set to cancel at period end.
        """
        _require_id("preview_upcoming_invoice", subscription_id)
        response = self._http.request(
            Request(method="GET", path=_action_path(subscription_id, "upcoming"))
        )
        invoice = response["data"].get("invoice")
        if invoice is None:
            return None
        return SubscriptionInvoicePreview.model_validate(invoice)

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[Subscription]:
        """Iterate every subscription matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[Subscription], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(
                    method="GET",
                    path="/subscriptions",
                    query=_encode_list_params(page_params),
                )
            )
            response = ListSubscriptionsResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncSubscriptionsService:
    """Async subscriptions service — bound as ``client.subscriptions`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListSubscriptionsResponse:
        body = await self._http.request(
            Request(method="GET", path="/subscriptions", query=_encode_list_params(params))
        )
        return ListSubscriptionsResponse.model_validate(body)

    async def retrieve(
        self, subscription_id: str, params: RetrieveParams | None = None
    ) -> Subscription:
        _require_id("retrieve", subscription_id)
        body = await self._http.request(
            Request(
                method="GET",
                path=_path_for(subscription_id),
                query=_encode_retrieve_params(params),
            )
        )
        return Subscription.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Subscription:
        if body is None:
            raise ValidationError(
                code="missing_body", message="subscriptions.create: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path="/subscriptions", body=payload)
        )
        return Subscription.model_validate(response["data"])

    async def update(self, subscription_id: str, body: UpdateBody) -> UpdateSubscriptionResult:
        _require_id("update", subscription_id)
        if body is None:
            raise ValidationError(
                code="missing_body", message="subscriptions.update: body must be non-None"
            )
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_path_for(subscription_id), body=payload)
        )
        return _build_update_result(response)

    async def activate(self, subscription_id: str) -> Subscription:
        _require_id("activate", subscription_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "activate"), body={})
        )
        return Subscription.model_validate(response["data"])

    async def cancel(self, subscription_id: str, body: CancelBody | None = None) -> Subscription:
        _require_id("cancel", subscription_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = await self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "cancel"), body=payload)
        )
        return Subscription.model_validate(response["data"])

    async def cancel_immediately(
        self, subscription_id: str, body: CancelImmediatelyBody | None = None
    ) -> Subscription:
        _require_id("cancel_immediately", subscription_id)
        payload = body.model_dump(by_alias=True, exclude_none=True) if body is not None else {}
        response = await self._http.request(
            Request(
                method="POST",
                path=_action_path(subscription_id, "cancel-immediately"),
                body=payload,
            )
        )
        return Subscription.model_validate(response["data"])

    async def mark_unpaid(self, subscription_id: str) -> Subscription:
        _require_id("mark_unpaid", subscription_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "mark-unpaid"), body={})
        )
        return Subscription.model_validate(response["data"])

    async def bill(self, subscription_id: str) -> BillSubscriptionResult:
        _require_id("bill", subscription_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "bill"), body={})
        )
        return _build_bill_result(response)

    async def renew(self, subscription_id: str) -> RenewSubscriptionResult:
        _require_id("renew", subscription_id)
        response = await self._http.request(
            Request(method="POST", path=_action_path(subscription_id, "renew"), body={})
        )
        return _build_renew_result(response)

    async def preview_upcoming_invoice(
        self, subscription_id: str
    ) -> SubscriptionInvoicePreview | None:
        _require_id("preview_upcoming_invoice", subscription_id)
        response = await self._http.request(
            Request(method="GET", path=_action_path(subscription_id, "upcoming"))
        )
        invoice = response["data"].get("invoice")
        if invoice is None:
            return None
        return SubscriptionInvoicePreview.model_validate(invoice)

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[Subscription]:
        """Async iterate every subscription matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[Subscription], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(
                    method="GET",
                    path="/subscriptions",
                    query=_encode_list_params(page_params),
                )
            )
            response = ListSubscriptionsResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
