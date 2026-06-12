"""Sync and async forms services.

Both services share the same wire shape and validation logic; the only
difference is which HTTP client they call.
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from urllib.parse import quote

from threecommon._core.http_client import Request
from threecommon.errors.classes import ValidationError
from threecommon.forms.types import (
    AddElementBody,
    AddLogicRuleBody,
    CreateBody,
    DeleteElementResult,
    DuplicateBody,
    EnableOtherOptionBody,
    Form,
    FormElement,
    FormSummary,
    ListFormsResponse,
    ListParams,
    MoveElementBody,
    UpdateBody,
    UpdateElementBody,
)
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


def _require(method: str, label: str, value: str) -> None:
    if not value:
        msg = f"forms.{method}: {label} must be a non-empty string"
        raise ValidationError(code="missing_id", message=msg)


def _require_body(method: str, body: object) -> None:
    if body is None:
        raise ValidationError(code="missing_body", message=f"forms.{method}: body must be non-None")


def _forms_path() -> str:
    return "/forms"


def _form_path(form_id: str) -> str:
    return f"/forms/{quote(form_id, safe='')}"


def _elements_path(form_id: str) -> str:
    return f"{_form_path(form_id)}/elements"


def _element_path(form_id: str, element_id: str) -> str:
    return f"{_elements_path(form_id)}/{quote(element_id, safe='')}"


def _logic_rules_path(form_id: str, element_id: str) -> str:
    return f"{_element_path(form_id, element_id)}/logic-rules"


# ---------------
# Sync
# ---------------


class FormsService:
    """Sync forms service - bound as ``client.forms`` on [ThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: HTTPClient) -> None:
        self._http = http

    def list(self, params: ListParams | None = None) -> ListFormsResponse:
        """List the host's forms (one page).

        For full iteration use [list_auto_paginate][FormsService.list_auto_paginate].
        """
        body = self._http.request(
            Request(method="GET", path=_forms_path(), query=_encode_list_params(params))
        )
        return ListFormsResponse.model_validate(body)

    def retrieve(self, form_id: str) -> Form:
        """Retrieve a single form by id."""
        _require("retrieve", "id", form_id)
        body = self._http.request(Request(method="GET", path=_form_path(form_id)))
        return Form.model_validate(body["data"])

    def create(self, body: CreateBody) -> Form:
        """Create a new form."""
        _require_body("create", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(Request(method="POST", path=_forms_path(), body=payload))
        return Form.model_validate(response["data"])

    def update(self, form_id: str, body: UpdateBody) -> Form:
        """Edit a form's settings. Only the supplied fields are changed."""
        _require("update", "id", form_id)
        _require_body("update", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_form_path(form_id), body=payload)
        )
        return Form.model_validate(response["data"])

    def duplicate(self, form_id: str, body: DuplicateBody | None = None) -> Form:
        """Duplicate a form, optionally overriding its name and status."""
        _require("duplicate", "id", form_id)
        payload = (body or DuplicateBody()).model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=f"{_form_path(form_id)}/duplicate", body=payload)
        )
        return Form.model_validate(response["data"])

    def add_element(self, form_id: str, body: AddElementBody) -> FormElement:
        """Add an element (question) to a form."""
        _require("add_element", "id", form_id)
        _require_body("add_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=_elements_path(form_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    def update_element(self, form_id: str, element_id: str, body: UpdateElementBody) -> FormElement:
        """Edit an element. Only the supplied fields are changed."""
        _require("update_element", "id", form_id)
        _require("update_element", "element_id", element_id)
        _require_body("update_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="PATCH", path=_element_path(form_id, element_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    def delete_element(self, form_id: str, element_id: str) -> DeleteElementResult:
        """Permanently remove an element. Echoes the removed element's id."""
        _require("delete_element", "id", form_id)
        _require("delete_element", "element_id", element_id)
        response = self._http.request(
            Request(method="DELETE", path=_element_path(form_id, element_id))
        )
        return DeleteElementResult.model_validate(response["data"])

    def move_element(self, form_id: str, element_id: str, body: MoveElementBody) -> Form:
        """Move an element to a new position. Returns the updated form."""
        _require("move_element", "id", form_id)
        _require("move_element", "element_id", element_id)
        _require_body("move_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(
                method="PUT",
                path=f"{_element_path(form_id, element_id)}/position",
                body=payload,
            )
        )
        return Form.model_validate(response["data"])

    def add_logic_rule(self, form_id: str, element_id: str, body: AddLogicRuleBody) -> FormElement:
        """Add a conditional-logic rule to a selection element."""
        _require("add_logic_rule", "id", form_id)
        _require("add_logic_rule", "element_id", element_id)
        _require_body("add_logic_rule", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(method="POST", path=_logic_rules_path(form_id, element_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    def remove_logic_rule(
        self, form_id: str, element_id: str, target_element_id: str
    ) -> FormElement:
        """Remove the logic rule revealing ``target_element_id``."""
        _require("remove_logic_rule", "id", form_id)
        _require("remove_logic_rule", "element_id", element_id)
        _require("remove_logic_rule", "target_element_id", target_element_id)
        path = f"{_logic_rules_path(form_id, element_id)}/{quote(target_element_id, safe='')}"
        response = self._http.request(Request(method="DELETE", path=path))
        return FormElement.model_validate(response["data"])

    def enable_other_option(
        self, form_id: str, element_id: str, body: EnableOtherOptionBody
    ) -> FormElement:
        """Enable the free-text "Other" option on a selection element."""
        _require("enable_other_option", "id", form_id)
        _require("enable_other_option", "element_id", element_id)
        _require_body("enable_other_option", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = self._http.request(
            Request(
                method="PUT",
                path=f"{_element_path(form_id, element_id)}/other-option",
                body=payload,
            )
        )
        return FormElement.model_validate(response["data"])

    def disable_other_option(self, form_id: str, element_id: str) -> FormElement:
        """Disable the free-text "Other" option on a selection element."""
        _require("disable_other_option", "id", form_id)
        _require("disable_other_option", "element_id", element_id)
        response = self._http.request(
            Request(method="DELETE", path=f"{_element_path(form_id, element_id)}/other-option")
        )
        return FormElement.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> Iter[FormSummary]:
        """Iterate every form matching ``params``, paging automatically."""
        start_page = params.page if params is not None and params.page is not None else 0

        def fetch(page: int) -> tuple[list[FormSummary], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = self._http.request(
                Request(method="GET", path=_forms_path(), query=_encode_list_params(page_params))
            )
            response = ListFormsResponse.model_validate(body)
            return response.data, response.has_more

        return Iter(fetch_page=fetch, start_page=start_page)


# ---------------
# Async
# ---------------


class AsyncFormsService:
    """Async forms service - bound as ``client.forms`` on [AsyncThreeCommon]."""

    __slots__ = ("_http",)

    def __init__(self, http: AsyncHTTPClient) -> None:
        self._http = http

    async def list(self, params: ListParams | None = None) -> ListFormsResponse:
        body = await self._http.request(
            Request(method="GET", path=_forms_path(), query=_encode_list_params(params))
        )
        return ListFormsResponse.model_validate(body)

    async def retrieve(self, form_id: str) -> Form:
        _require("retrieve", "id", form_id)
        body = await self._http.request(Request(method="GET", path=_form_path(form_id)))
        return Form.model_validate(body["data"])

    async def create(self, body: CreateBody) -> Form:
        _require_body("create", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_forms_path(), body=payload)
        )
        return Form.model_validate(response["data"])

    async def update(self, form_id: str, body: UpdateBody) -> Form:
        _require("update", "id", form_id)
        _require_body("update", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_form_path(form_id), body=payload)
        )
        return Form.model_validate(response["data"])

    async def duplicate(self, form_id: str, body: DuplicateBody | None = None) -> Form:
        _require("duplicate", "id", form_id)
        payload = (body or DuplicateBody()).model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=f"{_form_path(form_id)}/duplicate", body=payload)
        )
        return Form.model_validate(response["data"])

    async def add_element(self, form_id: str, body: AddElementBody) -> FormElement:
        _require("add_element", "id", form_id)
        _require_body("add_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_elements_path(form_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    async def update_element(
        self, form_id: str, element_id: str, body: UpdateElementBody
    ) -> FormElement:
        _require("update_element", "id", form_id)
        _require("update_element", "element_id", element_id)
        _require_body("update_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="PATCH", path=_element_path(form_id, element_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    async def delete_element(self, form_id: str, element_id: str) -> DeleteElementResult:
        _require("delete_element", "id", form_id)
        _require("delete_element", "element_id", element_id)
        response = await self._http.request(
            Request(method="DELETE", path=_element_path(form_id, element_id))
        )
        return DeleteElementResult.model_validate(response["data"])

    async def move_element(self, form_id: str, element_id: str, body: MoveElementBody) -> Form:
        _require("move_element", "id", form_id)
        _require("move_element", "element_id", element_id)
        _require_body("move_element", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(
                method="PUT",
                path=f"{_element_path(form_id, element_id)}/position",
                body=payload,
            )
        )
        return Form.model_validate(response["data"])

    async def add_logic_rule(
        self, form_id: str, element_id: str, body: AddLogicRuleBody
    ) -> FormElement:
        _require("add_logic_rule", "id", form_id)
        _require("add_logic_rule", "element_id", element_id)
        _require_body("add_logic_rule", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(method="POST", path=_logic_rules_path(form_id, element_id), body=payload)
        )
        return FormElement.model_validate(response["data"])

    async def remove_logic_rule(
        self, form_id: str, element_id: str, target_element_id: str
    ) -> FormElement:
        _require("remove_logic_rule", "id", form_id)
        _require("remove_logic_rule", "element_id", element_id)
        _require("remove_logic_rule", "target_element_id", target_element_id)
        path = f"{_logic_rules_path(form_id, element_id)}/{quote(target_element_id, safe='')}"
        response = await self._http.request(Request(method="DELETE", path=path))
        return FormElement.model_validate(response["data"])

    async def enable_other_option(
        self, form_id: str, element_id: str, body: EnableOtherOptionBody
    ) -> FormElement:
        _require("enable_other_option", "id", form_id)
        _require("enable_other_option", "element_id", element_id)
        _require_body("enable_other_option", body)
        payload = body.model_dump(by_alias=True, exclude_none=True)
        response = await self._http.request(
            Request(
                method="PUT",
                path=f"{_element_path(form_id, element_id)}/other-option",
                body=payload,
            )
        )
        return FormElement.model_validate(response["data"])

    async def disable_other_option(self, form_id: str, element_id: str) -> FormElement:
        _require("disable_other_option", "id", form_id)
        _require("disable_other_option", "element_id", element_id)
        response = await self._http.request(
            Request(method="DELETE", path=f"{_element_path(form_id, element_id)}/other-option")
        )
        return FormElement.model_validate(response["data"])

    def list_auto_paginate(self, params: ListParams | None = None) -> AsyncIter[FormSummary]:
        """Async iterate every form matching ``params``."""
        start_page = params.page if params is not None and params.page is not None else 0
        http = self._http

        async def fetch(page: int) -> tuple[list[FormSummary], bool]:
            page_params = (
                params.model_copy(update={"page": page})
                if params is not None
                else ListParams(page=page)
            )
            body = await http.request(
                Request(method="GET", path=_forms_path(), query=_encode_list_params(page_params))
            )
            response = ListFormsResponse.model_validate(body)
            return response.data, response.has_more

        return AsyncIter(fetch_page=fetch, start_page=start_page)
