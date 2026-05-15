"""Events-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.events import ListParams, RetrieveParams, UpdateBody


def _camel(s: str) -> str:
    parts = s.split("_")
    return parts[0] + "".join(p.title() for p in parts[1:])


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    for key in ("status", "page", "page_size", "search", "fields", "filters"):
        # Accept both snake_case and camelCase from YAML scenarios.
        for src in (key, _camel(key)):
            if src in args:
                payload[key] = args[src]
                break
    if "pageSize" in args and "page_size" not in payload:
        payload["page_size"] = args["pageSize"]
    return ListParams.model_validate(payload) if payload else None


def dispatch_sync(client: ThreeCommon, method: str, args: dict[str, Any]) -> Any:  # noqa: ANN401
    if method == "list":
        return client.events.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return client.events.retrieve(args["id"], params)
    if method == "update":
        body_raw = args.get("body", {}) or {}
        return client.events.update(args["id"], UpdateBody(**body_raw))
    if method == "listAutoPaginate":
        return list(client.events.list_auto_paginate(build_list_params(args)))
    pytest.fail(f"unsupported event method: {method}")


async def dispatch_async(client: AsyncThreeCommon, method: str, args: dict[str, Any]) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.events.list(build_list_params(args))
    if method == "retrieve":
        params_raw = args.get("params")
        params = RetrieveParams(fields=params_raw["fields"]) if params_raw else None
        return await client.events.retrieve(args["id"], params)
    if method == "update":
        body_raw = args.get("body", {}) or {}
        return await client.events.update(args["id"], UpdateBody(**body_raw))
    if method == "listAutoPaginate":
        return [ev async for ev in client.events.list_auto_paginate(build_list_params(args))]
    pytest.fail(f"unsupported event method: {method}")
