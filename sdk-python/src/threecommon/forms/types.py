"""Public types for the forms resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions; the element-authoring request bodies
use ``extra="allow"`` so element-type-specific fields the SDK doesn't yet
enumerate still reach the API.
"""

from __future__ import annotations

from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field

#: The type of form: determines whether the form drives an event checkout flow
#: (``order``) or is used on its own (``standalone``).
FormType = Literal["standalone", "order"]

#: Lifecycle status of a form.
#:
#: * ``draft``: not reachable at its URL, may be hidden from the organizer's tables
#: * ``active``: reachable at its URL and visible everywhere
#: * ``archived``: effectively deleted, but restorable from the forms dashboard
FormStatus = Literal["draft", "active", "archived"]

#: How wide the form's submit button renders.
SubmitButtonWidth = Literal["auto", "fill"]

#: How the form's submit button is aligned.
SubmitButtonAlign = Literal["left", "center"]

#: The kind of a form element. Question types capture an answer; the two
#: ``Static`` types render fixed content. The two ``Other`` variants add a
#: free-text escape hatch to a selection question.
ElementType = Literal[
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

#: Operator applied to the options referenced by a selection logic group.
LogicOperator = Literal["all_of", "any_of", "none_of"]

#: Comparison applied by a Yes/No logic group's ``value``.
LogicSelectionType = Literal["is", "is_not"]

#: File categories a ``File`` question may accept.
FileCategory = Literal["images", "documents", "data", "audio", "video"]

#: For order forms, the section an element is moved into by ``move_element``.
MoveSection = Literal["buyer", "ticket-holder"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class FormSummary(_BaseModel):
    """Compact form projection returned by ``list``."""

    id: str
    name: str
    num_elements: int = Field(serialization_alias="numElements", validation_alias="numElements")
    type: FormType
    status: FormStatus


class FormColumn(_BaseModel):
    """One column in a form-layout row. Points at an element by index."""

    element_index: int = Field(serialization_alias="elementIndex", validation_alias="elementIndex")
    width_fraction: float = Field(
        serialization_alias="widthFraction", validation_alias="widthFraction"
    )


class FormRow(_BaseModel):
    """One row in a form's layout. Holds one or more columns."""

    columns: list[FormColumn]


class LogicGroup(_BaseModel):
    """A conditional logic group attached to an element.

    Selection questions populate ``revealed_element_index`` + ``option_indices``
    + ``operator``; Yes/No questions populate ``selection_type`` + ``value``.
    """

    revealed_element_index: int | None = Field(
        default=None,
        serialization_alias="revealedElementIndex",
        validation_alias="revealedElementIndex",
    )
    option_indices: list[int] | None = Field(
        default=None, serialization_alias="optionIndices", validation_alias="optionIndices"
    )
    operator: LogicOperator | None = None
    selection_type: LogicSelectionType | None = Field(
        default=None, serialization_alias="selectionType", validation_alias="selectionType"
    )
    value: bool | None = None


class Element(_BaseModel):
    """A single form element (a question or a static block).

    The fields present depend on ``type``; everything except ``type`` is
    optional here so one model can describe every element kind. Unknown
    server-side fields are dropped (``extra="ignore"``).
    """

    id: str | None = None
    prompt: str | None = None
    prompt_hidden: bool | None = Field(
        default=None, serialization_alias="promptHidden", validation_alias="promptHidden"
    )
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    type: ElementType
    required: bool | None = None
    property_id: str | None = Field(
        default=None, serialization_alias="propertyId", validation_alias="propertyId"
    )
    property_data: dict[str, Any] | None = Field(
        default=None, serialization_alias="propertyData", validation_alias="propertyData"
    )
    contact_field: str | None = Field(
        default=None, serialization_alias="contactField", validation_alias="contactField"
    )
    placeholder: str | None = None
    options: list[str] | None = None
    dropdown: bool | None = None
    logic_groups: list[LogicGroup] | None = Field(
        default=None, serialization_alias="logicGroups", validation_alias="logicGroups"
    )
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )
    min_choices: int | None = Field(
        default=None, serialization_alias="minChoices", validation_alias="minChoices"
    )
    max_choices: int | None = Field(
        default=None, serialization_alias="maxChoices", validation_alias="maxChoices"
    )
    min: str | None = None
    max: str | None = None
    accept: list[FileCategory] | None = None
    content: str | None = None
    image_url: str | None = Field(
        default=None, serialization_alias="imageUrl", validation_alias="imageUrl"
    )
    image_width: float | None = Field(
        default=None, serialization_alias="imageWidth", validation_alias="imageWidth"
    )


class Form(_BaseModel):
    """Full form returned by ``retrieve``, ``create``, ``update``, ``duplicate``,
    and ``move_element``."""

    id: str
    name: str
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    owner_id: str = Field(serialization_alias="ownerId", validation_alias="ownerId")
    status: FormStatus
    rows: list[FormRow] = Field(default_factory=list)
    submit_button_text: str = Field(
        serialization_alias="submitButtonText", validation_alias="submitButtonText"
    )
    submit_button_width: SubmitButtonWidth = Field(
        serialization_alias="submitButtonWidth", validation_alias="submitButtonWidth"
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None, serialization_alias="submitButtonAlign", validation_alias="submitButtonAlign"
    )
    type: FormType
    elements: list[Element] = Field(default_factory=list)


class ListFormsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/forms``."""

    data: list[FormSummary]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class DeleteElementResult(_BaseModel):
    """Result shape returned by ``delete_element``. Echoes the removed element id."""

    deleted_element_id: str = Field(
        serialization_alias="deletedElementId", validation_alias="deletedElementId"
    )


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/forms``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    type: FormType | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/forms``.

    ``type`` selects the form kind. ``name`` is required by the API but is
    optional here, so a create call that omits it surfaces the server's
    ``validation_error`` rather than failing locally.
    """

    type: FormType
    name: str | None = None
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    status: FormStatus | None = None
    submit_button_text: str | None = Field(
        default=None, serialization_alias="submitButtonText", validation_alias="submitButtonText"
    )
    submit_button_width: SubmitButtonWidth | None = Field(
        default=None, serialization_alias="submitButtonWidth", validation_alias="submitButtonWidth"
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None, serialization_alias="submitButtonAlign", validation_alias="submitButtonAlign"
    )


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/forms/{id}``. All fields optional."""

    name: str | None = None
    name_hidden: bool | None = Field(
        default=None, serialization_alias="nameHidden", validation_alias="nameHidden"
    )
    status: FormStatus | None = None
    submit_button_text: str | None = Field(
        default=None, serialization_alias="submitButtonText", validation_alias="submitButtonText"
    )
    submit_button_width: SubmitButtonWidth | None = Field(
        default=None, serialization_alias="submitButtonWidth", validation_alias="submitButtonWidth"
    )
    submit_button_align: SubmitButtonAlign | None = Field(
        default=None, serialization_alias="submitButtonAlign", validation_alias="submitButtonAlign"
    )


class DuplicateBody(_BaseModel):
    """Body accepted by ``POST /v1/forms/{id}/duplicate``."""

    name: str | None = None
    status: FormStatus | None = None


class AddElementBody(_BaseModel):
    """Body accepted by ``POST /v1/forms/{id}/elements``.

    ``type`` selects the element kind; the remaining fields apply per kind.
    Unmodeled element-type-specific fields are forwarded as-is (``extra="allow"``).
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
        str_strip_whitespace=False,
    )

    type: ElementType
    prompt: str | None = None
    prompt_hidden: bool | None = Field(
        default=None, serialization_alias="promptHidden", validation_alias="promptHidden"
    )
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    required: bool | None = None
    property_id: str | None = Field(
        default=None, serialization_alias="propertyId", validation_alias="propertyId"
    )
    contact_field: str | None = Field(
        default=None, serialization_alias="contactField", validation_alias="contactField"
    )
    placeholder: str | None = None
    options: list[str] | None = None
    dropdown: bool | None = None
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )
    min_choices: int | None = Field(
        default=None, serialization_alias="minChoices", validation_alias="minChoices"
    )
    max_choices: int | None = Field(
        default=None, serialization_alias="maxChoices", validation_alias="maxChoices"
    )
    min: str | None = None
    max: str | None = None
    accept: list[FileCategory] | None = None
    content: str | None = None
    image_url: str | None = Field(
        default=None, serialization_alias="imageUrl", validation_alias="imageUrl"
    )
    image_width: float | None = Field(
        default=None, serialization_alias="imageWidth", validation_alias="imageWidth"
    )


class UpdateElementBody(_BaseModel):
    """Body accepted by ``PATCH /v1/forms/{id}/elements/{elementId}``.

    Every field is optional; only the fields you set are changed. Unmodeled
    element-type-specific fields are forwarded as-is (``extra="allow"``).
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
        str_strip_whitespace=False,
    )

    prompt: str | None = None
    prompt_hidden: bool | None = Field(
        default=None, serialization_alias="promptHidden", validation_alias="promptHidden"
    )
    helper_text: str | None = Field(
        default=None, serialization_alias="helperText", validation_alias="helperText"
    )
    placeholder: str | None = None
    required: bool | None = None
    property_id: str | None = Field(
        default=None, serialization_alias="propertyId", validation_alias="propertyId"
    )
    contact_field: str | None = Field(
        default=None, serialization_alias="contactField", validation_alias="contactField"
    )
    options: list[str] | None = None
    dropdown: bool | None = None
    other_prompt: str | None = Field(
        default=None, serialization_alias="otherPrompt", validation_alias="otherPrompt"
    )
    min_choices: int | None = Field(
        default=None, serialization_alias="minChoices", validation_alias="minChoices"
    )
    max_choices: int | None = Field(
        default=None, serialization_alias="maxChoices", validation_alias="maxChoices"
    )
    min: str | None = None
    max: str | None = None
    accept: list[FileCategory] | None = None
    content: str | None = None
    image_url: str | None = Field(
        default=None, serialization_alias="imageUrl", validation_alias="imageUrl"
    )
    image_width: float | None = Field(
        default=None, serialization_alias="imageWidth", validation_alias="imageWidth"
    )


class MoveElementBody(_BaseModel):
    """Body accepted by ``PUT /v1/forms/{id}/elements/{elementId}/position``."""

    position: int
    section: MoveSection | None = None


class EnableOtherOptionBody(_BaseModel):
    """Body accepted by ``PUT /v1/forms/{id}/elements/{elementId}/other-option``."""

    other_prompt: str = Field(serialization_alias="otherPrompt", validation_alias="otherPrompt")


class LogicCondition(_BaseModel):
    """The condition half of an ``add_logic_rule`` body.

    Selection questions set ``option_indices`` + ``operator``; Yes/No questions
    set ``selection_type`` + ``value``.
    """

    option_indices: list[int] | None = Field(
        default=None, serialization_alias="optionIndices", validation_alias="optionIndices"
    )
    operator: LogicOperator | None = None
    selection_type: LogicSelectionType | None = Field(
        default=None, serialization_alias="selectionType", validation_alias="selectionType"
    )
    value: bool | None = None


class AddLogicRuleBody(_BaseModel):
    """Body accepted by ``POST /v1/forms/{id}/elements/{elementId}/logic-rules``."""

    revealed_element_id: str = Field(
        serialization_alias="revealedElementId", validation_alias="revealedElementId"
    )
    condition: LogicCondition
