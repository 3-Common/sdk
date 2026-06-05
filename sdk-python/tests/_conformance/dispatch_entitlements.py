"""Entitlements-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.entitlements import (
    ConsumeBody,
    GrantBody,
    ListParams,
    LookupParams,
    RetrieveParams,
)

_LIST_PARAM_MAP = {
    "page": "page",
    "pageSize": "page_size",
    "contactId": "contact_id",
    "featureKey": "feature_key",
    "minBalance": "min_balance",
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


def build_lookup_params(args: dict[str, Any]) -> LookupParams:
    return LookupParams(
        contact_id=args.get("contactId", args.get("contact_id")),
        feature_key=args.get("featureKey", args.get("feature_key")),
        fields=args.get("fields"),
    )


def build_grant_body(raw: dict[str, Any]) -> GrantBody:
    return GrantBody(
        contact_id=raw.get("contactId", raw.get("contact_id")),
        feature_key=raw.get("featureKey", raw.get("feature_key")),
        amount=raw.get("amount"),
        grant_id=raw.get("grantId", raw.get("grant_id")),
        metadata=raw.get("metadata"),
    )


def build_consume_body(raw: dict[str, Any]) -> ConsumeBody:
    return ConsumeBody(
        contact_id=raw.get("contactId", raw.get("contact_id")),
        feature_key=raw.get("featureKey", raw.get("feature_key")),
        amount=raw.get("amount"),
        reason=raw.get("reason"),
    )


def dispatch_sync(  # noqa: PLR0911
    client: ThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return client.entitlements.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return client.entitlements.retrieve(args["id"], params)
    if method == "lookup":
        return client.entitlements.lookup(build_lookup_params(args))
    if method == "grant":
        return client.entitlements.grant(build_grant_body(args.get("body") or {}))
    if method == "consume":
        return client.entitlements.consume(build_consume_body(args.get("body") or {}))
    if method == "listAutoPaginate":
        return list(client.entitlements.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported entitlement method: {method}")


async def dispatch_async(  # noqa: PLR0911
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.entitlements.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return await client.entitlements.retrieve(args["id"], params)
    if method == "lookup":
        return await client.entitlements.lookup(build_lookup_params(args))
    if method == "grant":
        return await client.entitlements.grant(build_grant_body(args.get("body") or {}))
    if method == "consume":
        return await client.entitlements.consume(build_consume_body(args.get("body") or {}))
    if method == "listAutoPaginate":
        return [e async for e in client.entitlements.list_auto_paginate(build_list_params(args))]
    pytest.fail(f"unsupported entitlement method: {method}")
