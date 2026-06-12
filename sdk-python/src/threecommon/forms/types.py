"""Public types for the forms resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). Most response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions. ``FormElement`` is the exception: form
elements carry a wide, element-type-dependent set of fields, so both the
element request bodies and ``FormElement`` itself use ``extra="allow"`` to
preserve anything the SDK doesn't model explicitly rather than drop it.
"""

from __future__ import annotations

from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field

#: The two kinds of form. ``standalone`` forms collect submissions on their own;
#: ``order`` forms are attached to a checkout flow.
FormType = Literal["standalone", "order"]

#: Lifecycle status of a form.
FormStatus = Literal["draft", "active", "archived"]

#: How wide the submit button renders.
SubmitButtonWidth = Literal["auto", "fill"]

#: How the submit button is horizontally aligned.
SubmitButtonAlign = Literal["left", "center"]

#: Which section of an ``order`` form an element lives in when it is moved.
ElementSection = Literal["buyer", "ticket-holder"]

#: The element kinds accepted by ``add_element``. Mirrors the OpenAPI union.
#: The response-side ``FormElement.type`` is deliberately ``str`` instead, so
#: element types added server-side never break response parsing.
FormElementType = Literal[
    "Text",
    "Multi-line Text",
    "Select One",
    'Select One or "Other"',
    "Select Multiple",
    'Select Multiple with "Other"',
    "Yes/No",
    "Date",
    "File",
    "Email",
    "Phone",
    "Static Text",
    "Static Image",
]

#: How a selection-question logic condition combines its ``option_indices``.
LogicOperator = Literal["all_of", "any_of", "none_of"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class _ElementBodyBase(BaseModel):
    """Base config for element request bodies.

    Form elements accept a wide, element-type-dependent set of fields. The
    commonly used ones are modeled explicitly below; ``extra="allow"`` forwards
    any additional fields verbatim so callers are never blocked by the SDK.
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
        str_strip_whitespace=False,
    )


# -----------------------------------------------------------------------------
# Response models
# -----------------------------------------------------------------------------


class FormSummary(_BaseModel):
    """A form in the compact projection returned by ``list``.

    All fields are required: the list endpoint always returns every one of
    them.
    """

    id: str
    name: str
    num_elements: int = Field(serialization_alias="numElements", validation_alias="numElements")
    type: FormType
    status: FormStatus


class FormColumn(_BaseModel):
    """One column in a form-layout row. Points at an element by its index in the
    form's ``elements`` array and how much of the row's width it occupies."""

    element_index: int = Field(serialization_alias="elementIndex", validation_alias="elementIndex")
    width_fraction: float = Field(
        serialization_alias="widthFraction", validation_alias="widthFraction"
    )


class FormRow(_BaseModel):
    """One row in a form's layout. ``columns`` lists the elements shown in the
    row, left to right, with their relative widths."""

    columns: list[FormColumn]


class FormElement(BaseModel):
    """A single element (question or static element) on a form.

    The ``type`` field is a free-form label (e.g. ``"Text"``, ``"Select One"``,
    ``'Select One or "Other"'``). The commonly used fields are modeled
    explicitly; element-type-specific fields (``dropdown``, ``logicGroups``,
    ``propertyData``, etc.) are preserved verbatim via ``extra="allow"`` rather
    than dropped. ``required`` is optional because several element types
    (``Select Multiple`` and the static text/image elements) omit it.
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
        str_strip_whitespace=False,
    )

    id: str
    type: str
    prompt: str
    required: bool | None = None
    prompt_hidden: bool | None = Field(
        default=None, serialization_alias="promptHidden", validation_alias="promptHidden"
    )
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    placeholder: str | None = None
    options: list[Any] | None = None
    property_id: str | None = Field(
        default=None, serialization_alias="propertyId", validation_alias="propertyId"
    )
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )


class Form(_BaseModel):
    """A form in the full projection returned by ``retrieve``, ``create``,
    ``update``, ``duplicate``, and ``move_element`` - including its ``elements``
    and layout ``rows``. ``attendee_rows_start`` is populated on ``order``
    forms only: it is the index in ``rows`` where the ticket-holder section
    begins."""

    id: str
    name: str
    owner_id: str = Field(serialization_alias="ownerId", validation_alias="ownerId")
    type: FormType
    status: FormStatus
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    submit_button_text: str | None = Field(
        default=None,
        serialization_alias="submitButtonText",
        validation_alias="submitButtonText",
    )
    submit_button_width: SubmitButtonWidth | None = Field(
        default=None,
        serialization_alias="submitButtonWidth",
        validation_alias="submitButtonWidth",
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None,
        serialization_alias="submitButtonAlign",
        validation_alias="submitButtonAlign",
    )
    elements: list[FormElement] | None = None
    rows: list[FormRow] | None = None
    attendee_rows_start: int | None = Field(
        default=None,
        serialization_alias="attendeeRowsStart",
        validation_alias="attendeeRowsStart",
    )


class DeleteElementResult(_BaseModel):
    """Result shape returned by ``delete_element``. Echoes the removed element id."""

    deleted_element_id: str = Field(
        serialization_alias="deletedElementId", validation_alias="deletedElementId"
    )


class ListFormsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/forms``."""

    data: list[FormSummary]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


# -----------------------------------------------------------------------------
# Request models
# -----------------------------------------------------------------------------


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/forms``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    type: FormType | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/forms``.

    ``type`` is fixed at creation time and cannot be changed afterwards.
    """

    name: str
    type: FormType
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    status: FormStatus | None = None
    submit_button_text: str | None = Field(
        default=None,
        serialization_alias="submitButtonText",
        validation_alias="submitButtonText",
    )
    submit_button_width: SubmitButtonWidth | None = Field(
        default=None,
        serialization_alias="submitButtonWidth",
        validation_alias="submitButtonWidth",
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None,
        serialization_alias="submitButtonAlign",
        validation_alias="submitButtonAlign",
    )


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/forms/{id}``. Every field is optional;
    only the supplied fields are changed. ``name_hidden`` and
    ``submit_button_align`` accept an explicit ``None`` (sent as JSON ``null``)
    to clear the setting server-side."""

    name: str | None = None
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    status: FormStatus | None = None
    submit_button_text: str | None = Field(
        default=None,
        serialization_alias="submitButtonText",
        validation_alias="submitButtonText",
    )
    submit_button_width: SubmitButtonWidth | None = Field(
        default=None,
        serialization_alias="submitButtonWidth",
        validation_alias="submitButtonWidth",
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None,
        serialization_alias="submitButtonAlign",
        validation_alias="submitButtonAlign",
    )


class DuplicateBody(_BaseModel):
    """Body accepted by ``POST /v1/forms/{id}/duplicate``."""

    name: str | None = None
    status: FormStatus | None = None


class AddElementBody(_ElementBodyBase):
    """Body accepted by ``POST /v1/forms/{id}/elements``.

    ``type`` selects the element kind; the remaining fields are the commonly
    used options. Any element-type-specific field not modeled here is forwarded
    verbatim (``extra="allow"``).
    """

    type: FormElementType
    prompt: str | None = None
    required: bool | None = None
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    placeholder: str | None = None
    options: list[Any] | None = None
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )


class UpdateElementBody(_ElementBodyBase):
    """Body accepted by ``PATCH /v1/forms/{id}/elements/{elementId}``.

    Every field is optional; only the supplied fields are changed. Fields not
    modeled here are forwarded verbatim (``extra="allow"``). Most fields accept
    an explicit ``None`` (sent as JSON ``null``) to clear the setting
    server-side.
    """

    prompt: str | None = None
    required: bool | None = None
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    placeholder: str | None = None
    options: list[Any] | None = None
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )


class MoveElementBody(_BaseModel):
    """Body accepted by ``PUT /v1/forms/{id}/elements/{elementId}/position``."""

    position: int
    section: ElementSection | None = None


class EnableOtherOptionBody(_BaseModel):
    """Body accepted by ``PUT /v1/forms/{id}/elements/{elementId}/other-option``."""

    other_prompt: str = Field(serialization_alias="otherPrompt", validation_alias="otherPrompt")


class SelectionLogicCondition(_BaseModel):
    """Logic condition for selection questions.

    ``option_indices`` selects which of the source element's options trigger
    the rule (an index of ``-1`` means the "Other" option, where applicable);
    ``operator`` describes how the respondent's selections must relate to them.
    """

    option_indices: list[int] = Field(
        serialization_alias="optionIndices", validation_alias="optionIndices"
    )
    operator: LogicOperator


class YesNoLogicCondition(_BaseModel):
    """Logic condition for Yes/No questions.

    ``value`` is compared to the question's answer; ``selection_type`` decides
    whether the answer must equal it (``"is"``) or differ from it
    (``"is_not"``). An unanswered question is treated differently from an
    explicit ``False`` answer.
    """

    selection_type: Literal["is", "is_not"] = Field(
        serialization_alias="selectionType", validation_alias="selectionType"
    )
    value: bool


#: Either condition shape accepted by ``add_logic_rule``: selection questions
#: take a ``SelectionLogicCondition``, Yes/No questions take a
#: ``YesNoLogicCondition``.
LogicCondition = SelectionLogicCondition | YesNoLogicCondition


class AddLogicRuleBody(_BaseModel):
    """Body accepted by ``POST /v1/forms/{id}/elements/{elementId}/logic-rules``.

    The ``condition`` shape depends on the source element:
    ``SelectionLogicCondition`` for selection questions,
    ``YesNoLogicCondition`` for Yes/No questions.
    """

    revealed_element_id: str = Field(
        serialization_alias="revealedElementId", validation_alias="revealedElementId"
    )
    condition: LogicCondition
