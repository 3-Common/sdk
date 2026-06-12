"""Forms service tests - sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pydantic import ValidationError as PydanticValidationError
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
    FormSummary,
    ListParams,
    MoveElementBody,
    SelectionLogicCondition,
    UpdateBody,
    UpdateElementBody,
    YesNoLogicCondition,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


SAMPLE_FORM = {
    "id": "frm_123",
    "name": "Registration",
    "ownerId": "hst_1",
    "type": "standalone",
    "status": "active",
    "submitButtonText": "Sign up",
}

SAMPLE_SUMMARY = {
    "id": "frm_a",
    "name": "Newsletter Signup",
    "numElements": 3,
    "type": "standalone",
    "status": "active",
}

SAMPLE_ELEMENT = {
    "id": "elm_1",
    "prompt": "What is your name?",
    "type": "Text",
    "required": True,
}


# ---------------
# Sync forms
# ---------------


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?type=standalone&pageSize=10",
        json={"data": [SAMPLE_SUMMARY], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.forms.list(ListParams(type="standalone", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "frm_a"
    assert result.data[0].num_elements == 3
    assert result.has_more is False


def test_form_summary_requires_wire_fields() -> None:
    # The list endpoint always returns these; a missing one should fail fast
    # rather than silently default to None.
    with pytest.raises(PydanticValidationError):
        FormSummary.model_validate({"id": "frm_1", "name": "First"})


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.forms.list()
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.retrieve("frm_123")
    assert form.id == "frm_123"
    assert form.name == "Registration"
    assert form.owner_id == "hst_1"
    assert form.submit_button_text == "Sign up"


def test_retrieve_preserves_elements_and_layout(httpx_mock: HTTPXMock) -> None:
    full_form = {
        **SAMPLE_FORM,
        "elements": [
            {
                "id": "elm_1",
                "type": "Select One",
                "prompt": "Pick one",
                "required": True,
                "options": ["A", "B"],
                "dropdown": True,
                "logicGroups": [
                    {"revealedElementIndex": 1, "optionIndices": [0], "operator": "any_of"}
                ],
            },
            {
                "id": "elm_2",
                "type": "Static Image",
                "prompt": "Banner",
                "src": "https://x.test/a.png",
            },
        ],
        "rows": [{"columns": [{"elementIndex": 0, "widthFraction": 1.0}]}],
    }
    httpx_mock.add_response(url="http://test.local/v1/forms/frm_123", json={"data": full_form})
    with _make_sync() as c:
        form = c.forms.retrieve("frm_123")

    assert form.elements is not None
    assert [e.id for e in form.elements] == ["elm_1", "elm_2"]

    select = form.elements[0]
    assert select.type == "Select One"
    assert select.options == ["A", "B"]
    # element-type-specific fields are preserved verbatim, not dropped.
    assert select.model_extra is not None
    assert select.model_extra["dropdown"] is True
    assert select.model_extra["logicGroups"][0]["operator"] == "any_of"

    # a static element (no `required`) still parses; `required` defaults to None.
    image = form.elements[1]
    assert image.prompt == "Banner"
    assert image.required is None
    assert image.model_extra is not None
    assert image.model_extra["src"] == "https://x.test/a.png"

    # layout rows are modeled with snake_case access.
    assert form.rows is not None
    assert form.rows[0].columns[0].element_index == 0
    assert form.rows[0].columns[0].width_fraction == 1.0


def test_retrieve_order_form_keeps_attendee_rows_start(httpx_mock: HTTPXMock) -> None:
    order_form = {**SAMPLE_FORM, "type": "order", "attendeeRowsStart": 2}
    httpx_mock.add_response(url="http://test.local/v1/forms/frm_123", json={"data": order_form})
    with _make_sync() as c:
        form = c.forms.retrieve("frm_123")
    assert form.attendee_rows_start == 2


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
        match_json={"name": "Registration", "type": "standalone"},
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.create(CreateBody(name="Registration", type="standalone"))
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
        json={"error": {"code": "validation_failed", "message": "bad name"}},
    )
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.create(CreateBody(name="Bad", type="standalone"))
    assert exc.value.code == "validation_failed"


def test_create_rejects_invalid_type() -> None:
    # `type` is a closed two-value enum; typos fail client-side.
    with pytest.raises(PydanticValidationError):
        CreateBody.model_validate({"name": "Bad", "type": "not-a-form-type"})


def test_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        method="PATCH",
        match_json={"name": "Updated", "status": "active"},
        json={"data": {**SAMPLE_FORM, "name": "Updated"}},
    )
    with _make_sync() as c:
        form = c.forms.update("frm_123", UpdateBody(name="Updated", status="active"))
    assert form.name == "Updated"


def test_update_sends_explicit_null_to_clear(httpx_mock: HTTPXMock) -> None:
    # Fields explicitly set to None go over the wire as JSON null (the API's
    # "clear this setting" signal); unset fields are omitted entirely.
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123",
        method="PATCH",
        match_json={"submitButtonAlign": None},
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        c.forms.update("frm_123", UpdateBody(submit_button_align=None))


def test_update_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.update("", UpdateBody(name="x"))


def test_update_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.update("frm_123", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_duplicate_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/duplicate",
        method="POST",
        match_json={"name": "Registration (Copy)", "status": "draft"},
        json={"data": {**SAMPLE_FORM, "id": "frm_copy", "name": "Registration (Copy)"}},
    )
    with _make_sync() as c:
        copy = c.forms.duplicate(
            "frm_123", DuplicateBody(name="Registration (Copy)", status="draft")
        )
    assert copy.id == "frm_copy"


def test_duplicate_without_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/duplicate",
        method="POST",
        json={"data": {**SAMPLE_FORM, "id": "frm_copy"}},
    )
    with _make_sync() as c:
        copy = c.forms.duplicate("frm_123")
    assert copy.id == "frm_copy"


def test_duplicate_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.duplicate("")


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
    assert element.id == "elm_1"
    assert element.prompt == "What is your name?"


def test_add_element_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.add_element("", AddElementBody(type="Text"))


def test_add_element_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.add_element("frm_123", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_add_element_rejects_unknown_type() -> None:
    # `type` is pinned to the element kinds in the OpenAPI union.
    with pytest.raises(PydanticValidationError):
        AddElementBody.model_validate({"type": "Carousel"})


def test_update_element_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1",
        method="PATCH",
        match_json={"prompt": "What is your full name?", "required": False},
        json={"data": {**SAMPLE_ELEMENT, "prompt": "What is your full name?", "required": False}},
    )
    with _make_sync() as c:
        element = c.forms.update_element(
            "frm_123",
            "elm_1",
            UpdateElementBody(prompt="What is your full name?", required=False),
        )
    assert element.prompt == "What is your full name?"
    assert element.required is False


def test_update_element_sends_explicit_null_to_clear(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1",
        method="PATCH",
        match_json={"helperText": None, "placeholder": None},
        json={"data": SAMPLE_ELEMENT},
    )
    with _make_sync() as c:
        c.forms.update_element(
            "frm_123", "elm_1", UpdateElementBody(helper_text=None, placeholder=None)
        )


def test_update_element_validates_element_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.update_element("frm_123", "", UpdateElementBody(prompt="x"))


def test_update_element_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_missing",
        method="PATCH",
        status_code=404,
        json={"error": {"code": "not_found", "message": "gone"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.forms.update_element("frm_123", "elm_missing", UpdateElementBody(prompt="x"))


def test_delete_element_returns_id(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1",
        method="DELETE",
        json={"data": {"deletedElementId": "elm_1"}},
    )
    with _make_sync() as c:
        result = c.forms.delete_element("frm_123", "elm_1")
    assert result.deleted_element_id == "elm_1"


def test_delete_element_validates_ids() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError):
            c.forms.delete_element("", "elm_1")
        with pytest.raises(ValidationError):
            c.forms.delete_element("frm_123", "")


def test_move_element_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/position",
        method="PUT",
        match_json={"position": 2},
        json={"data": SAMPLE_FORM},
    )
    with _make_sync() as c:
        form = c.forms.move_element("frm_123", "elm_1", MoveElementBody(position=2))
    assert form.id == "frm_123"


def test_move_element_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.forms.move_element("frm_123", "elm_1", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_add_logic_rule_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules",
        method="POST",
        match_json={
            "revealedElementId": "elm_2",
            "condition": {"optionIndices": [0], "operator": "any_of"},
        },
        json={"data": {**SAMPLE_ELEMENT, "type": "Select One"}},
    )
    with _make_sync() as c:
        element = c.forms.add_logic_rule(
            "frm_123",
            "elm_1",
            AddLogicRuleBody(
                revealed_element_id="elm_2",
                condition=SelectionLogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
    assert element.id == "elm_1"
    assert element.type == "Select One"


def test_add_logic_rule_yes_no_condition(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules",
        method="POST",
        match_json={
            "revealedElementId": "elm_2",
            "condition": {"selectionType": "is", "value": True},
        },
        json={"data": {**SAMPLE_ELEMENT, "type": "Yes/No"}},
    )
    with _make_sync() as c:
        element = c.forms.add_logic_rule(
            "frm_123",
            "elm_1",
            AddLogicRuleBody(
                revealed_element_id="elm_2",
                condition=YesNoLogicCondition(selection_type="is", value=True),
            ),
        )
    assert element.type == "Yes/No"


def test_add_logic_rule_body_accepts_both_condition_shapes() -> None:
    # Wire-shaped (camelCase) input, as the conformance harness supplies it.
    selection = AddLogicRuleBody.model_validate(
        {"revealedElementId": "elm_2", "condition": {"optionIndices": [0], "operator": "any_of"}}
    )
    assert isinstance(selection.condition, SelectionLogicCondition)

    yes_no = AddLogicRuleBody.model_validate(
        {"revealedElementId": "elm_2", "condition": {"selectionType": "is_not", "value": False}}
    )
    assert isinstance(yes_no.condition, YesNoLogicCondition)


def test_add_logic_rule_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.add_logic_rule("frm_123", "elm_1", None)  # type: ignore[arg-type]


def test_remove_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules/elm_2",
        method="DELETE",
        json={"data": {**SAMPLE_ELEMENT, "type": "Select One"}},
    )
    with _make_sync() as c:
        element = c.forms.remove_logic_rule("frm_123", "elm_1", "elm_2")
    assert element.id == "elm_1"


def test_remove_logic_rule_validates_target() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.remove_logic_rule("frm_123", "elm_1", "")


def test_enable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/other-option",
        method="PUT",
        match_json={"otherPrompt": "Other (please specify)"},
        json={
            "data": {
                **SAMPLE_ELEMENT,
                "type": 'Select One or "Other"',
                "otherPrompt": "Other (please specify)",
            }
        },
    )
    with _make_sync() as c:
        element = c.forms.enable_other_option(
            "frm_123", "elm_1", EnableOtherOptionBody(other_prompt="Other (please specify)")
        )
    assert element.other_prompt == "Other (please specify)"


def test_enable_other_option_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.enable_other_option("frm_123", "elm_1", None)  # type: ignore[arg-type]


def test_disable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/other-option",
        method="DELETE",
        json={"data": {**SAMPLE_ELEMENT, "type": "Select One"}},
    )
    with _make_sync() as c:
        element = c.forms.disable_other_option("frm_123", "elm_1")
    assert element.id == "elm_1"


def test_disable_other_option_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.forms.disable_other_option("frm_123", "")


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?type=standalone&page=0",
        json={
            "data": [
                {**SAMPLE_SUMMARY, "id": "frm_1", "name": "First"},
                {**SAMPLE_SUMMARY, "id": "frm_2", "name": "Second"},
            ],
            "hasMore": True,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/forms?type=standalone&page=1",
        json={"data": [{**SAMPLE_SUMMARY, "id": "frm_3", "name": "Third"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [form.id for form in c.forms.list_auto_paginate(ListParams(type="standalone"))]
    assert ids == ["frm_1", "frm_2", "frm_3"]


def test_list_auto_paginate_no_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?page=0",
        json={"data": [], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [form.id for form in c.forms.list_auto_paginate()]
    assert ids == []


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
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms",
        method="POST",
        json={"data": SAMPLE_FORM},
    )
    async with _make_async() as c:
        form = await c.forms.create(CreateBody(name="Registration", type="standalone"))
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
        copy = await c.forms.duplicate("frm_123", DuplicateBody(name="Copy"))
    assert copy.id == "frm_copy"


@pytest.mark.asyncio
async def test_async_duplicate_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.duplicate("")


@pytest.mark.asyncio
async def test_async_add_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements",
        method="POST",
        json={"data": SAMPLE_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.add_element("frm_123", AddElementBody(type="Text"))
    assert element.id == "elm_1"


@pytest.mark.asyncio
async def test_async_add_element_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.add_element("frm_123", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_update_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1",
        method="PATCH",
        json={"data": SAMPLE_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.update_element("frm_123", "elm_1", UpdateElementBody(required=True))
    assert element.id == "elm_1"


@pytest.mark.asyncio
async def test_async_update_element_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.update_element("", "elm_1", UpdateElementBody())


@pytest.mark.asyncio
async def test_async_delete_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1",
        method="DELETE",
        json={"data": {"deletedElementId": "elm_1"}},
    )
    async with _make_async() as c:
        result = await c.forms.delete_element("frm_123", "elm_1")
    assert result.deleted_element_id == "elm_1"


@pytest.mark.asyncio
async def test_async_delete_element_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.delete_element("frm_123", "")


@pytest.mark.asyncio
async def test_async_move_element(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/position",
        method="PUT",
        json={"data": SAMPLE_FORM},
    )
    async with _make_async() as c:
        form = await c.forms.move_element("frm_123", "elm_1", MoveElementBody(position=1))
    assert form.id == "frm_123"


@pytest.mark.asyncio
async def test_async_move_element_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.move_element("frm_123", "elm_1", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_add_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules",
        method="POST",
        json={"data": {**SAMPLE_ELEMENT, "type": "Select One"}},
    )
    async with _make_async() as c:
        element = await c.forms.add_logic_rule(
            "frm_123",
            "elm_1",
            AddLogicRuleBody(
                revealed_element_id="elm_2",
                condition=SelectionLogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
    assert element.id == "elm_1"


@pytest.mark.asyncio
async def test_async_add_logic_rule_yes_no_condition(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules",
        method="POST",
        match_json={
            "revealedElementId": "elm_2",
            "condition": {"selectionType": "is", "value": True},
        },
        json={"data": {**SAMPLE_ELEMENT, "type": "Yes/No"}},
    )
    async with _make_async() as c:
        element = await c.forms.add_logic_rule(
            "frm_123",
            "elm_1",
            AddLogicRuleBody(
                revealed_element_id="elm_2",
                condition=YesNoLogicCondition(selection_type="is", value=True),
            ),
        )
    assert element.type == "Yes/No"


@pytest.mark.asyncio
async def test_async_add_logic_rule_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.add_logic_rule("frm_123", "elm_1", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_remove_logic_rule(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/logic-rules/elm_2",
        method="DELETE",
        json={"data": {**SAMPLE_ELEMENT, "type": "Select One"}},
    )
    async with _make_async() as c:
        element = await c.forms.remove_logic_rule("frm_123", "elm_1", "elm_2")
    assert element.id == "elm_1"


@pytest.mark.asyncio
async def test_async_remove_logic_rule_validates_target() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.remove_logic_rule("frm_123", "elm_1", "")


@pytest.mark.asyncio
async def test_async_enable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/other-option",
        method="PUT",
        json={"data": {**SAMPLE_ELEMENT, "otherPrompt": "Other"}},
    )
    async with _make_async() as c:
        element = await c.forms.enable_other_option(
            "frm_123", "elm_1", EnableOtherOptionBody(other_prompt="Other")
        )
    assert element.other_prompt == "Other"


@pytest.mark.asyncio
async def test_async_enable_other_option_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.enable_other_option("frm_123", "elm_1", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_disable_other_option(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms/frm_123/elements/elm_1/other-option",
        method="DELETE",
        json={"data": SAMPLE_ELEMENT},
    )
    async with _make_async() as c:
        element = await c.forms.disable_other_option("frm_123", "elm_1")
    assert element.id == "elm_1"


@pytest.mark.asyncio
async def test_async_disable_other_option_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.forms.disable_other_option("", "elm_1")


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/forms?page=0",
        json={
            "data": [
                {**SAMPLE_SUMMARY, "id": "a", "name": "A"},
                {**SAMPLE_SUMMARY, "id": "b", "name": "B"},
            ],
            "hasMore": True,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/forms?page=1",
        json={"data": [{**SAMPLE_SUMMARY, "id": "c", "name": "C"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [form.id async for form in c.forms.list_auto_paginate()]
    assert ids == ["a", "b", "c"]
