"""Properties-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.properties import CreateBody, ListParams, UpdateBody


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    # The scenario args use the wire (camelCase) keys; populate_by_name +
    # validation aliases let model_validate accept them directly.
    return ListParams.model_validate(args) if args else None


def dispatch_sync(  # noqa: PLR0911
    client: ThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return client.properties.list(build_list_params(args))
    if method == "retrieve":
        return client.properties.retrieve(args["id"])
    if method == "create":
        return client.properties.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return client.properties.update(
            args["id"], UpdateBody.model_validate(args.get("body") or {})
        )
    if method == "listAutoPaginate":
        return list(client.properties.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported properties method: {method}")


async def dispatch_async(  # noqa: PLR0911
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.properties.list(build_list_params(args))
    if method == "retrieve":
        return await client.properties.retrieve(args["id"])
    if method == "create":
        return await client.properties.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return await client.properties.update(
            args["id"], UpdateBody.model_validate(args.get("body") or {})
        )
    if method == "listAutoPaginate":
        return [p async for p in client.properties.list_auto_paginate(build_list_params(args))]
    pytest.fail(f"unsupported properties method: {method}")
