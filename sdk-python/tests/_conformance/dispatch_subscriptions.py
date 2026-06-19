"""Subscriptions-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.subscriptions import (
    CancelBody,
    CancelImmediatelyBody,
    CreateBody,
    CreateBodyItem,
    ListParams,
    RetrieveParams,
    SubscriptionTaxId,
    UpdateBody,
)

_LIST_PARAM_MAP = {
    "page": "page",
    "pageSize": "page_size",
    "status": "status",
    "contactId": "contact_id",
    "priceId": "price_id",
    "fields": "fields",
}


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    for camel, snake in _LIST_PARAM_MAP.items():
        if camel in args:
            payload[snake] = args[camel]
        elif snake in args:
            payload[snake] = args[snake]
    return ListParams.model_validate(payload) if payload else None


def _tax_ids(raw: Any) -> list[SubscriptionTaxId] | None:  # noqa: ANN401
    if raw is None:
        return None
    return [SubscriptionTaxId(type=t["type"], value=t["value"]) for t in raw]


def _items(raw: Any) -> list[CreateBodyItem] | None:  # noqa: ANN401
    if raw is None:
        return None
    return [
        CreateBodyItem(
            price_id=i.get("priceId", i.get("price_id")),
            quantity=i.get("quantity"),
        )
        for i in raw
    ]


def build_create_body(raw: dict[str, Any]) -> CreateBody:
    return CreateBody(
        price_id=raw.get("priceId", raw.get("price_id")),
        quantity=raw.get("quantity"),
        items=_items(raw.get("items")),
        contact_id=raw.get("contactId", raw.get("contact_id")),
        customer_email=raw.get("customerEmail", raw.get("customer_email")),
        trial_days=raw.get("trialDays", raw.get("trial_days")),
        billing_cycle_anchor=raw.get("billingCycleAnchor", raw.get("billing_cycle_anchor")),
        cancel_at=raw.get("cancelAt", raw.get("cancel_at")),
        dunning_enabled=raw.get("dunningEnabled", raw.get("dunning_enabled")),
        notes=raw.get("notes"),
        tax_ids=_tax_ids(raw.get("taxIds") or raw.get("tax_ids")),
        auto_charge=raw.get("autoCharge", raw.get("auto_charge")),
        payment_due_days=raw.get("paymentDueDays", raw.get("payment_due_days")),
        tax_rate=raw.get("taxRate", raw.get("tax_rate")),
        metadata=raw.get("metadata"),
    )


def build_update_body(raw: dict[str, Any]) -> UpdateBody:
    return UpdateBody(
        price_id=raw.get("priceId", raw.get("price_id")),
        quantity=raw.get("quantity"),
        notes=raw.get("notes"),
        tax_ids=_tax_ids(raw.get("taxIds") or raw.get("tax_ids")),
        tax_rate=raw.get("taxRate", raw.get("tax_rate")),
        auto_charge=raw.get("autoCharge", raw.get("auto_charge")),
        dunning_enabled=raw.get("dunningEnabled", raw.get("dunning_enabled")),
        payment_due_days=raw.get("paymentDueDays", raw.get("payment_due_days")),
    )


def dispatch_sync(  # noqa: PLR0911, PLR0912
    client: ThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return client.subscriptions.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return client.subscriptions.retrieve(args["id"], params)
    if method == "create":
        return client.subscriptions.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return client.subscriptions.update(args["id"], build_update_body(args.get("body") or {}))
    if method == "retrieveManageUrl":
        return client.subscriptions.retrieve_manage_url(args["id"])
    if method == "activate":
        return client.subscriptions.activate(args["id"])
    if method == "cancel":
        body_raw = args.get("body")
        cancel_body = CancelBody(reason=body_raw.get("reason")) if body_raw else None
        return client.subscriptions.cancel(args["id"], cancel_body)
    if method == "cancelImmediately":
        body_raw = args.get("body")
        cancel_imm_body = CancelImmediatelyBody(reason=body_raw.get("reason")) if body_raw else None
        return client.subscriptions.cancel_immediately(args["id"], cancel_imm_body)
    if method == "markUnpaid":
        return client.subscriptions.mark_unpaid(args["id"])
    if method == "bill":
        return client.subscriptions.bill(args["id"])
    if method == "renew":
        return client.subscriptions.renew(args["id"])
    if method == "previewUpcomingInvoice":
        return client.subscriptions.preview_upcoming_invoice(args["id"])
    if method == "listAutoPaginate":
        return list(client.subscriptions.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported subscription method: {method}")


async def dispatch_async(  # noqa: PLR0911, PLR0912
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.subscriptions.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return await client.subscriptions.retrieve(args["id"], params)
    if method == "create":
        return await client.subscriptions.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return await client.subscriptions.update(
            args["id"], build_update_body(args.get("body") or {})
        )
    if method == "retrieveManageUrl":
        return await client.subscriptions.retrieve_manage_url(args["id"])
    if method == "activate":
        return await client.subscriptions.activate(args["id"])
    if method == "cancel":
        body_raw = args.get("body")
        cancel_body = CancelBody(reason=body_raw.get("reason")) if body_raw else None
        return await client.subscriptions.cancel(args["id"], cancel_body)
    if method == "cancelImmediately":
        body_raw = args.get("body")
        cancel_imm_body = CancelImmediatelyBody(reason=body_raw.get("reason")) if body_raw else None
        return await client.subscriptions.cancel_immediately(args["id"], cancel_imm_body)
    if method == "markUnpaid":
        return await client.subscriptions.mark_unpaid(args["id"])
    if method == "bill":
        return await client.subscriptions.bill(args["id"])
    if method == "renew":
        return await client.subscriptions.renew(args["id"])
    if method == "previewUpcomingInvoice":
        return await client.subscriptions.preview_upcoming_invoice(args["id"])
    if method == "listAutoPaginate":
        return [s async for s in client.subscriptions.list_auto_paginate(build_list_params(args))]
    pytest.fail(f"unsupported subscription method: {method}")
