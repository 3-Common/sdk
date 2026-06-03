"""Prices-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.prices import CreateBody, ListParams, RetrieveParams, UpdateBody


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    # The scenario args use the wire (camelCase) keys; populate_by_name +
    # validation aliases let model_validate accept them directly.
    return ListParams.model_validate(args) if args else None


def _retrieve_params(args: dict[str, Any]) -> RetrieveParams | None:
    params_raw = args.get("params")
    return RetrieveParams(fields=params_raw["fields"]) if params_raw else None


def dispatch_sync(  # noqa: PLR0911
    client: ThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return client.prices.list(build_list_params(args))
    if method == "retrieve":
        return client.prices.retrieve(args["id"], _retrieve_params(args))
    if method == "create":
        return client.prices.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return client.prices.update(args["id"], UpdateBody.model_validate(args.get("body") or {}))
    if method == "archive":
        return client.prices.archive(args["id"])
    if method == "unarchive":
        return client.prices.unarchive(args["id"])
    if method == "listAutoPaginate":
        return list(client.prices.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported price method: {method}")


async def dispatch_async(  # noqa: PLR0911
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.prices.list(build_list_params(args))
    if method == "retrieve":
        return await client.prices.retrieve(args["id"], _retrieve_params(args))
    if method == "create":
        return await client.prices.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return await client.prices.update(
            args["id"], UpdateBody.model_validate(args.get("body") or {})
        )
    if method == "archive":
        return await client.prices.archive(args["id"])
    if method == "unarchive":
        return await client.prices.unarchive(args["id"])
    if method == "listAutoPaginate":
        return [p async for p in client.prices.list_auto_paginate(build_list_params(args))]
    pytest.fail(f"unsupported price method: {method}")
