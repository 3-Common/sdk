"""Invoices-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.invoices import (
    CreateBody,
    InvoiceLineItem,
    ListParams,
    PaymentBody,
    RetrieveParams,
    UpdateBody,
    VoidBody,
)


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    mapping = {
        "page": "page",
        "pageSize": "page_size",
        "status": "status",
        "customerId": "customer_id",
        "issuedAfter": "issued_after",
        "issuedBefore": "issued_before",
        "fields": "fields",
    }
    for camel, snake in mapping.items():
        if camel in args:
            payload[snake] = args[camel]
        elif snake in args:
            payload[snake] = args[snake]
    return ListParams.model_validate(payload) if payload else None


def _line_items(raw: Any) -> list[InvoiceLineItem]:  # noqa: ANN401
    items: list[InvoiceLineItem] = []
    for entry in raw or []:
        kw: dict[str, Any] = {
            "description": entry["description"],
            "quantity": entry["quantity"],
            "unit_amount": entry.get("unitAmount", entry.get("unit_amount")),
        }
        if "productId" in entry or "product_id" in entry:
            kw["product_id"] = entry.get("productId", entry.get("product_id"))
        if "taxAmount" in entry or "tax_amount" in entry:
            kw["tax_amount"] = entry.get("taxAmount", entry.get("tax_amount"))
        items.append(InvoiceLineItem(**kw))
    return items


def build_create_body(raw: dict[str, Any]) -> CreateBody:
    return CreateBody(
        customer_id=raw.get("customerId", raw.get("customer_id")),
        currency=raw["currency"],
        line_items=_line_items(raw.get("lineItems") or raw.get("line_items")),
        notes=raw.get("notes"),
        due_at=raw.get("dueAt", raw.get("due_at")),
        subscription_id=raw.get("subscriptionId", raw.get("subscription_id")),
        quote_id=raw.get("quoteId", raw.get("quote_id")),
    )


def build_update_body(raw: dict[str, Any]) -> UpdateBody:
    line_items_raw = raw.get("lineItems") or raw.get("line_items")
    return UpdateBody(
        customer_id=raw.get("customerId", raw.get("customer_id")),
        line_items=_line_items(line_items_raw) if line_items_raw is not None else None,
        notes=raw.get("notes"),
        due_at=raw.get("dueAt", raw.get("due_at")),
    )


def dispatch_sync(client: ThreeCommon, method: str, args: dict[str, Any]) -> Any:  # noqa: ANN401, PLR0911
    if method == "list":
        return client.invoices.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return client.invoices.retrieve(args["id"], params)
    if method == "create":
        return client.invoices.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return client.invoices.update(args["id"], build_update_body(args.get("body") or {}))
    if method == "finalize":
        return client.invoices.finalize(args["id"])
    if method == "void":
        body_raw = args.get("body")
        body = VoidBody(reason=body_raw.get("reason")) if body_raw else None
        return client.invoices.void(args["id"], body)
    if method == "recordPayment":
        body_raw = args.get("body") or {}
        return client.invoices.record_payment(
            args["id"],
            PaymentBody(
                payment=body_raw["payment"],
                idempotency_key=body_raw.get("idempotencyKey", body_raw.get("idempotency_key")),
                note=body_raw.get("note"),
            ),
        )
    if method == "listAutoPaginate":
        return list(client.invoices.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported invoice method: {method}")


async def dispatch_async(  # noqa: PLR0911
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.invoices.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return await client.invoices.retrieve(args["id"], params)
    if method == "create":
        return await client.invoices.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return await client.invoices.update(
            args["id"], build_update_body(args.get("body") or {})
        )
    if method == "finalize":
        return await client.invoices.finalize(args["id"])
    if method == "void":
        body_raw = args.get("body")
        body = VoidBody(reason=body_raw.get("reason")) if body_raw else None
        return await client.invoices.void(args["id"], body)
    if method == "recordPayment":
        body_raw = args.get("body") or {}
        return await client.invoices.record_payment(
            args["id"],
            PaymentBody(
                payment=body_raw["payment"],
                idempotency_key=body_raw.get("idempotencyKey", body_raw.get("idempotency_key")),
                note=body_raw.get("note"),
            ),
        )
    if method == "listAutoPaginate":
        return [
            inv async for inv in client.invoices.list_auto_paginate(build_list_params(args))
        ]
    pytest.fail(f"unsupported invoice method: {method}")
