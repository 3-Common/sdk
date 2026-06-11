"""Forms-resource dispatcher for the conformance harness."""

from __future__ import annotations

from typing import Any

import pytest

from threecommon import AsyncThreeCommon, ThreeCommon
from threecommon.forms import (
    AddElementBody,
    AddLogicRuleBody,
    CreateBody,
    DuplicateBody,
    EnableOtherOptionBody,
    ListParams,
    MoveElementBody,
    UpdateBody,
    UpdateElementBody,
)


def build_list_params(args: dict[str, Any]) -> ListParams | None:
    if not args:
        return None
    payload: dict[str, Any] = {}
    mapping = {"page": "page", "pageSize": "page_size", "type": "type"}
    for camel, snake in mapping.items():
        if camel in args:
            payload[snake] = args[camel]
        elif snake in args:
            payload[snake] = args[snake]
    return ListParams.model_validate(payload) if payload else None


def dispatch_sync(client: ThreeCommon, method: str, args: dict[str, Any]) -> Any:  # noqa: ANN401, PLR0911, PLR0912
    if method == "list":
        return client.forms.list(build_list_params(args))
    if method == "listAutoPaginate":
        return list(client.forms.list_auto_paginate(build_list_params(args)))
    if method == "retrieve":
        return client.forms.retrieve(args["id"])
    if method == "create":
        return client.forms.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return client.forms.update(args["id"], UpdateBody.model_validate(args.get("body") or {}))
    if method == "duplicate":
        return client.forms.duplicate(
            args["id"], DuplicateBody.model_validate(args.get("body") or {})
        )
    if method == "addElement":
        return client.forms.add_element(
            args["id"], AddElementBody.model_validate(args.get("body") or {})
        )
    if method == "updateElement":
        return client.forms.update_element(
            args["id"], args["elementId"], UpdateElementBody.model_validate(args.get("body") or {})
        )
    if method == "deleteElement":
        return client.forms.delete_element(args["id"], args["elementId"])
    if method == "moveElement":
        return client.forms.move_element(
            args["id"], args["elementId"], MoveElementBody.model_validate(args.get("body") or {})
        )
    if method == "addLogicRule":
        return client.forms.add_logic_rule(
            args["id"], args["elementId"], AddLogicRuleBody.model_validate(args.get("body") or {})
        )
    if method == "removeLogicRule":
        return client.forms.remove_logic_rule(
            args["id"], args["elementId"], args["targetElementId"]
        )
    if method == "enableOtherOption":
        return client.forms.enable_other_option(
            args["id"],
            args["elementId"],
            EnableOtherOptionBody.model_validate(args.get("body") or {}),
        )
    if method == "disableOtherOption":
        return client.forms.disable_other_option(args["id"], args["elementId"])
    pytest.fail(f"unsupported forms method: {method}")


async def dispatch_async(  # noqa: PLR0911, PLR0912
    client: AsyncThreeCommon, method: str, args: dict[str, Any]
) -> Any:  # noqa: ANN401
    if method == "list":
        return await client.forms.list(build_list_params(args))
    if method == "listAutoPaginate":
        return [f async for f in client.forms.list_auto_paginate(build_list_params(args))]
    if method == "retrieve":
        return await client.forms.retrieve(args["id"])
    if method == "create":
        return await client.forms.create(CreateBody.model_validate(args.get("body") or {}))
    if method == "update":
        return await client.forms.update(
            args["id"], UpdateBody.model_validate(args.get("body") or {})
        )
    if method == "duplicate":
        return await client.forms.duplicate(
            args["id"], DuplicateBody.model_validate(args.get("body") or {})
        )
    if method == "addElement":
        return await client.forms.add_element(
            args["id"], AddElementBody.model_validate(args.get("body") or {})
        )
    if method == "updateElement":
        return await client.forms.update_element(
            args["id"], args["elementId"], UpdateElementBody.model_validate(args.get("body") or {})
        )
    if method == "deleteElement":
        return await client.forms.delete_element(args["id"], args["elementId"])
    if method == "moveElement":
        return await client.forms.move_element(
            args["id"], args["elementId"], MoveElementBody.model_validate(args.get("body") or {})
        )
    if method == "addLogicRule":
        return await client.forms.add_logic_rule(
            args["id"], args["elementId"], AddLogicRuleBody.model_validate(args.get("body") or {})
        )
    if method == "removeLogicRule":
        return await client.forms.remove_logic_rule(
            args["id"], args["elementId"], args["targetElementId"]
        )
    if method == "enableOtherOption":
        return await client.forms.enable_other_option(
            args["id"],
            args["elementId"],
            EnableOtherOptionBody.model_validate(args.get("body") or {}),
        )
    if method == "disableOtherOption":
        return await client.forms.disable_other_option(args["id"], args["elementId"])
    pytest.fail(f"unsupported forms method: {method}")
