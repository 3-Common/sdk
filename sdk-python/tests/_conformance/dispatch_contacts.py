"""Contacts-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.contacts import (
    ActivityListParams,
    AttachPaymentMethodBody,
    BulkUpsertBody,
    BulkUpsertItem,
    ContactUpdate,
    CreateBody,
    ListParams,
    UpdateBody,
)


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    mapping = {
        "pageNumber": "page_number",
        "pageSize": "page_size",
        "sortField": "sort_field",
        "sortDirection": "sort_direction",
        "filter": "filter",
        "filters": "filters",
        "search": "search",
    }
    for camel, snake in mapping.items():
        if camel in args:
            payload[snake] = args[camel]
        elif snake in args:
            payload[snake] = args[snake]
    return ListParams.model_validate(payload) if payload else None


def build_activity_params(args: dict[str, Any] | None) -> ActivityListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    mapping = {
        "pageNumber": "page_number",
        "pageSize": "page_size",
        "filter": "filter",
        "sort": "sort",
    }
    for camel, snake in mapping.items():
        if camel in args:
            payload[snake] = args[camel]
        elif snake in args:
            payload[snake] = args[snake]
    return ActivityListParams.model_validate(payload) if payload else None


def build_create_body(raw: dict[str, Any]) -> CreateBody:
    return CreateBody(
        email=raw["email"],
        first_name=raw.get("firstName", raw.get("first_name")),
        last_name=raw.get("lastName", raw.get("last_name")),
        phone=raw.get("phone"),
    )


def build_update_body(raw: dict[str, Any]) -> UpdateBody:
    contact = raw["contact"]
    return UpdateBody(
        contact=ContactUpdate(
            first_name=contact.get("firstName", contact.get("first_name")),
            last_name=contact.get("lastName", contact.get("last_name")),
            email=contact["email"],
            phone=contact.get("phone"),
            status=contact["status"],
        ),
        merge_with=raw.get("mergeWith", raw.get("merge_with")),
        resolution=raw.get("resolution"),
    )


def build_bulk_upsert_body(raw: dict[str, Any]) -> BulkUpsertBody:
    items: list[BulkUpsertItem] = []
    for entry in raw.get("contacts") or []:
        items.append(
            BulkUpsertItem(
                email=entry["email"],
                first_name=entry.get("firstName", entry.get("first_name")),
                last_name=entry.get("lastName", entry.get("last_name")),
                phone=entry.get("phone"),
                status=entry.get("status"),
            )
        )
    return BulkUpsertBody(contacts=items)


def build_attach_payment_method_body(raw: dict[str, Any]) -> AttachPaymentMethodBody:
    return AttachPaymentMethodBody(
        setup_intent_id=raw.get("setupIntentId", raw.get("setup_intent_id")),
    )


def dispatch_sync(client: ThreeCommon, method: str, args: dict[str, Any]) -> Any:  # noqa: ANN401, PLR0911, PLR0912
    if method == "list":
        return client.contacts.list(build_list_params(args))
    if method == "count":
        return client.contacts.count()
    if method == "retrieve":
        return client.contacts.retrieve(args["id"])
    if method == "create":
        return client.contacts.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return client.contacts.update(args["id"], build_update_body(args.get("body") or {}))
    if method == "delete":
        return client.contacts.delete(args["id"])
    if method == "bulkUpsert":
        return client.contacts.bulk_upsert(build_bulk_upsert_body(args.get("body") or {}))
    if method == "listActivity":
        return client.contacts.list_activity(args["id"], build_activity_params(args.get("params")))
    if method == "listAutoPaginate":
        return list(client.contacts.list_auto_paginate(build_list_params(args)))
    if method == "listActivityAutoPaginate":
        return list(
            client.contacts.list_activity_auto_paginate(
                args["id"], build_activity_params(args.get("params"))
            )
        )
    if method == "retrievePaymentMethod":
        return client.contacts.retrieve_payment_method(args["id"])
    if method == "attachPaymentMethod":
        return client.contacts.attach_payment_method(
            args["id"], build_attach_payment_method_body(args.get("body") or {})
        )
    if method == "createPaymentMethodSetupIntent":
        return client.contacts.create_payment_method_setup_intent(args["id"])
    if method == "removePaymentMethod":
        return client.contacts.remove_payment_method(args["id"], args["methodId"])
    pytest.fail(f"unsupported contacts method: {method}")


async def dispatch_async(  # noqa: PLR0911, PLR0912
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.contacts.list(build_list_params(args))
    if method == "count":
        return await client.contacts.count()
    if method == "retrieve":
        return await client.contacts.retrieve(args["id"])
    if method == "create":
        return await client.contacts.create(build_create_body(args.get("body") or {}))
    if method == "update":
        return await client.contacts.update(args["id"], build_update_body(args.get("body") or {}))
    if method == "delete":
        return await client.contacts.delete(args["id"])
    if method == "bulkUpsert":
        return await client.contacts.bulk_upsert(build_bulk_upsert_body(args.get("body") or {}))
    if method == "listActivity":
        return await client.contacts.list_activity(
            args["id"], build_activity_params(args.get("params"))
        )
    if method == "listAutoPaginate":
        return [c async for c in client.contacts.list_auto_paginate(build_list_params(args))]
    if method == "listActivityAutoPaginate":
        return [
            a
            async for a in client.contacts.list_activity_auto_paginate(
                args["id"], build_activity_params(args.get("params"))
            )
        ]
    if method == "retrievePaymentMethod":
        return await client.contacts.retrieve_payment_method(args["id"])
    if method == "attachPaymentMethod":
        return await client.contacts.attach_payment_method(
            args["id"], build_attach_payment_method_body(args.get("body") or {})
        )
    if method == "createPaymentMethodSetupIntent":
        return await client.contacts.create_payment_method_setup_intent(args["id"])
    if method == "removePaymentMethod":
        return await client.contacts.remove_payment_method(args["id"], args["methodId"])
    pytest.fail(f"unsupported contacts method: {method}")
