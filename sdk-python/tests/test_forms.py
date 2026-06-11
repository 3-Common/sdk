"""Forms service tests - sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.forms import (
    AddElementBody,
    AddLogicRuleBody,
    CreateBody,
    DuplicateBody,
    EnableOtherOptionBody,
    ListParams,
    LogicCondition,
    MoveElementBody,
    UpdateBody,
    UpdateElementBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


SAMPLE_FORM = {
    "id": "frm_123",
    "name": "Customer survey",
    "ownerId": "hst_1",
    "status": "active",
    "type": "standalone",
    "submitButtonText": "Submit",
    "submitButtonWidth": "auto",
    "rows": [],
    "elements": [],
}

SAMPLE_ELEMENT = {
    "id": "elm_123",
    "prompt": "What is your name?",
    "type": "Text",
    "required": True,
}

SAMPLE_SELECT_ELEMENT = {
    "id": "elm_select",
    "prompt": "How did you hear about us?",
    "type": "Select One",
    "required": True,
    "options": ["Friend", "Social media"],
    "logicGroups": [
        {"revealedElementIndex": 3, "optionIndices": [0], "operator": "any_of"},
    ],
}

SAMPLE_SUMMARY = {
    "id": "frm_a",
    "name": "Customer survey",
    "numElements": 4,
    "type": "standalone",
    "status": "active",
}


# ---------------
# Sync forms
# ---------------


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?pageSize=10&type=standalone",
        json={"data": [SAMPLE_SUMMARY], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.forms.list(ListParams(type="standalone", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "frm_a"
    assert result.data[0].num_elements == 4
    assert result.has_more is False


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.forms.list()
    assert result.data == []


def test_list_empty_params_sends_no_query(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.forms.list(ListParams())
    assert result.has_more is False


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.retrieve("frm_123")
    assert form.id == "frm_123"
    assert form.name == "Customer survey"
    assert form.owner_id == "hst_1"


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.forms.retrieve("frm_missing")


def test_create_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        method="POST",
        match_json={"name": "Customer survey", "type": "standalone"},
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.create(CreateBody(name="Customer survey", type="standalone"))
    assert form.id == "frm_123"


def test_create_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_create_400_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        method="POST",
        status_code=400,
        json={"error": {"code": "validation_error", "message": "name is required"}},
    )
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.create(CreateBody(type="standalone"))


def test_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        method="PATCH",
        match_json={"name": "Renamed survey", "status": "active"},
        json={"data": {**SAMPLE_FORM, "name": "Renamed survey"}},
    )
    with _make_sync() as c:
        form = c.forms.update("frm_123", UpdateBody(name="Renamed survey", status="active"))
    assert form.name == "Renamed survey"


def test_update_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.update("", UpdateBody(name="x"))
    assert exc.value.code == "missing_id"


def test_update_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.update("frm_123", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_duplicate_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/duplicate",
        method="POST",
        match_json={"name": "Customer survey (copy)"},
        json={"data": {**SAMPLE_FORM, "id": "frm_copy", "name": "Customer survey (copy)"}},
    )
    with _make_sync() as c:
        copy = c.forms.duplicate("frm_123", DuplicateBody(name="Customer survey (copy)"))
    assert copy.id == "frm_copy"


def test_duplicate_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.duplicate("", DuplicateBody(name="x"))


def test_duplicate_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.duplicate("frm_123", None)  # type: ignore[arg-type]


def test_add_element_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements",
        method="POST",
        match_json={"type": "Text", "prompt": "What is your name?", "required": True},
        json={"data": SAMPLE_ELEMENT},
    )
    with _make_sync() as c:
        element = c.forms.add_element(
            "frm_123",
            AddElementBody(type="Text", prompt="What is your name?", required=True),
        )
    assert element.id == "elm_123"
    assert element.type == "Text"


def test_add_element_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.add_element("", AddElementBody(type="Text"))


def test_add_element_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.add_element("frm_123", None)  # type: ignore[arg-type]


def test_update_element_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123",
        method="PATCH",
        match_json={"prompt": "What is your full name?"},
        json={"data": {**SAMPLE_ELEMENT, "prompt": "What is your full name?"}},
    )
    with _make_sync() as c:
        element = c.forms.update_element(
            "frm_123", "elm_123", UpdateElementBody(prompt="What is your full name?")
        )
    assert element.prompt == "What is your full name?"


def test_update_element_validates_element_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.update_element("frm_123", "", UpdateElementBody(prompt="x"))


def test_update_element_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.update_element("frm_123", "elm_123", None)  # type: ignore[arg-type]


def test_delete_element_returns_id(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123",
        method="DELETE",
        json={"data": {"deletedElementId": "elm_123"}},
    )
    with _make_sync() as c:
        result = c.forms.delete_element("frm_123", "elm_123")
    assert result.deleted_element_id == "elm_123"


def test_delete_element_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.delete_element("", "elm_123")


def test_delete_element_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_missing",
        method="DELETE",
        status_code=404,
        json={"error": {"code": "not_found", "message": "gone"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.forms.delete_element("frm_123", "elm_missing")


def test_move_element_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123/position",
        method="PUT",
        match_json={"position": 2},
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.move_element("frm_123", "elm_123", MoveElementBody(position=2))
    assert form.id == "frm_123"


def test_move_element_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.move_element("frm_123", "elm_123", None)  # type: ignore[arg-type]


def test_add_logic_rule_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/logic-rules",
        method="POST",
        match_json={
            "revealedElementId": "elm_followup",
            "condition": {"optionIndices": [0], "operator": "any_of"},
        },
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    with _make_sync() as c:
        element = c.forms.add_logic_rule(
            "frm_123",
            "elm_select",
            AddLogicRuleBody(
                revealed_element_id="elm_followup",
                condition=LogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
    assert element.logic_groups is not None
    assert element.logic_groups[0].revealed_element_index == 3


def test_add_logic_rule_validates_element_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.add_logic_rule(
            "frm_123",
            "",
            AddLogicRuleBody(
                revealed_element_id="elm_followup",
                condition=LogicCondition(option_indices=[0], operator="any_of"),
            ),
        )


def test_remove_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/logic-rules/elm_followup",
        method="DELETE",
        json={"data": {**SAMPLE_SELECT_ELEMENT, "logicGroups": []}},
    )
    with _make_sync() as c:
        element = c.forms.remove_logic_rule("frm_123", "elm_select", "elm_followup")
    assert element.id == "elm_select"


def test_remove_logic_rule_validates_target_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.remove_logic_rule("frm_123", "elm_select", "")
    assert exc.value.code == "missing_id"


def test_enable_other_option_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/other-option",
        method="PUT",
        match_json={"otherPrompt": "Other (please specify)"},
        json={
            "data": {
                **SAMPLE_SELECT_ELEMENT,
                "type": 'Select One or "Other"',
                "otherPrompt": "Other (please specify)",
            }
        },
    )
    with _make_sync() as c:
        element = c.forms.enable_other_option(
            "frm_123", "elm_select", EnableOtherOptionBody(other_prompt="Other (please specify)")
        )
    assert element.other_prompt == "Other (please specify)"
    assert element.type == 'Select One or "Other"'


def test_enable_other_option_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.enable_other_option("frm_123", "elm_select", None)  # type: ignore[arg-type]


def test_disable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/other-option",
        method="DELETE",
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    with _make_sync() as c:
        element = c.forms.disable_other_option("frm_123", "elm_select")
    assert element.id == "elm_select"


def test_disable_other_option_validates_element_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.disable_other_option("frm_123", "")


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?type=standalone&page=0",
        json={
            "data": [{**SAMPLE_SUMMARY, "id": "frm_1"}, {**SAMPLE_SUMMARY, "id": "frm_2"}],
            "hasMore": True,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/forms?type=standalone&page=1",
        json={"data": [{**SAMPLE_SUMMARY, "id": "frm_3"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [f.id for f in c.forms.list_auto_paginate(ListParams(type="standalone"))]
    assert ids == ["frm_1", "frm_2", "frm_3"]


# ---------------
# Async forms
# ---------------


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        json={"data": [], "hasMore": False},
    )
    async with _make_async() as c:
        r = await c.forms.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_1",
        json={"data": {**SAMPLE_FORM, "id": "frm_1"}},
    )
    async with _make_async() as c:
        form = await c.forms.retrieve("frm_1")
    assert form.id == "frm_1"


@pytest.mark.asyncio
async def test_async_retrieve_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.retrieve("")


@pytest.mark.asyncio
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        method="POST",
        json={"data": SAMPLE_FORM},
    )
    async with _make_async() as c:
        form = await c.forms.create(CreateBody(name="Customer survey", type="standalone"))
    assert form.id == "frm_123"


@pytest.mark.asyncio
async def test_async_create_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.create(None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_update(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        method="PATCH",
        json={"data": SAMPLE_FORM},
    )
    async with _make_async() as c:
        form = await c.forms.update("frm_123", UpdateBody(status="active"))
    assert form.id == "frm_123"


@pytest.mark.asyncio
async def test_async_update_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.update("", UpdateBody(name="x"))


@pytest.mark.asyncio
async def test_async_update_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.update("frm_123", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_duplicate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/duplicate",
        method="POST",
        json={"data": {**SAMPLE_FORM, "id": "frm_copy"}},
    )
    async with _make_async() as c:
        copy = await c.forms.duplicate("frm_123", DuplicateBody(name="copy"))
    assert copy.id == "frm_copy"


@pytest.mark.asyncio
async def test_async_duplicate_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.duplicate("frm_123", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_add_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements",
        method="POST",
        json={"data": SAMPLE_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.add_element("frm_123", AddElementBody(type="Text", prompt="Name?"))
    assert element.id == "elm_123"


@pytest.mark.asyncio
async def test_async_add_element_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.add_element("", AddElementBody(type="Text"))


@pytest.mark.asyncio
async def test_async_update_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123",
        method="PATCH",
        json={"data": SAMPLE_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.update_element(
            "frm_123", "elm_123", UpdateElementBody(prompt="New?")
        )
    assert element.id == "elm_123"


@pytest.mark.asyncio
async def test_async_update_element_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.update_element("frm_123", "elm_123", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_delete_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123",
        method="DELETE",
        json={"data": {"deletedElementId": "elm_123"}},
    )
    async with _make_async() as c:
        result = await c.forms.delete_element("frm_123", "elm_123")
    assert result.deleted_element_id == "elm_123"


@pytest.mark.asyncio
async def test_async_delete_element_validates_element_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.delete_element("frm_123", "")


@pytest.mark.asyncio
async def test_async_move_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_123/position",
        method="PUT",
        json={"data": SAMPLE_FORM},
    )
    async with _make_async() as c:
        form = await c.forms.move_element(
            "frm_123", "elm_123", MoveElementBody(position=1, section="buyer")
        )
    assert form.id == "frm_123"


@pytest.mark.asyncio
async def test_async_move_element_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.move_element("", "elm_123", MoveElementBody(position=1))


@pytest.mark.asyncio
async def test_async_add_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/logic-rules",
        method="POST",
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.add_logic_rule(
            "frm_123",
            "elm_select",
            AddLogicRuleBody(
                revealed_element_id="elm_followup",
                condition=LogicCondition(selection_type="is", value=True),
            ),
        )
    assert element.id == "elm_select"


@pytest.mark.asyncio
async def test_async_add_logic_rule_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.add_logic_rule("frm_123", "elm_select", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_remove_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/logic-rules/elm_followup",
        method="DELETE",
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.remove_logic_rule("frm_123", "elm_select", "elm_followup")
    assert element.id == "elm_select"


@pytest.mark.asyncio
async def test_async_remove_logic_rule_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.remove_logic_rule("frm_123", "", "elm_followup")


@pytest.mark.asyncio
async def test_async_enable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/other-option",
        method="PUT",
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.enable_other_option(
            "frm_123", "elm_select", EnableOtherOptionBody(other_prompt="Other")
        )
    assert element.id == "elm_select"


@pytest.mark.asyncio
async def test_async_enable_other_option_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.enable_other_option("frm_123", "elm_select", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_disable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_select/other-option",
        method="DELETE",
        json={"data": SAMPLE_SELECT_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.disable_other_option("frm_123", "elm_select")
    assert element.id == "elm_select"


@pytest.mark.asyncio
async def test_async_disable_other_option_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.disable_other_option("", "elm_select")


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?page=0",
        json={
            "data": [{**SAMPLE_SUMMARY, "id": "a"}, {**SAMPLE_SUMMARY, "id": "b"}],
            "hasMore": True,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/forms?page=1",
        json={"data": [{**SAMPLE_SUMMARY, "id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [f.id async for f in c.forms.list_auto_paginate()]
    assert ids == ["a", "b", "c"]
