"""Public types for the events resource.

Hand-curated friendly aliases over the auto-generated OpenAPI types in
[threecommon._generated][]. Field names match the Python convention
(snake_case in the SDK; the wire payload preserves camelCase via Pydantic
field aliases).
"""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

#: Lifecycle status of an event. Surface-mirrors the API enum; future
#: server-side additions will arrive as raw strings until the SDK is updated.
EventStatus = Literal[
    "draft",
    "open",
    "closed",
    "unpublished",
    "cancelled",
    "postponed",
    "schedule",
]


class _BaseModel(BaseModel):
    """Base class with our default ``ConfigDict``.

    ``populate_by_name=True`` lets customers construct via either snake_case
    or camelCase. ``extra="ignore"`` keeps the SDK forward-compatible: new
    server-side fields don't break older SDKs.
    """

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class Event(_BaseModel):
    """One event as returned by the API.

    Optional fields are populated only when the server returned them — list
    responses with a ``fields`` filter omit unrequested values.
    """

    id: str
    name: str | None = None
    type: str | None = None
    schedule: str | None = None
    start: str | None = None  # ISO 8601
    status: EventStatus | None = None
    items_sold: int | None = Field(
        default=None, serialization_alias="itemsSold", validation_alias="itemsSold"
    )
    revenue_cents: int | None = Field(
        default=None, serialization_alias="revenueCents", validation_alias="revenueCents"
    )
    min_price_cents: int | None = Field(
        default=None, serialization_alias="minPriceCents", validation_alias="minPriceCents"
    )
    max_price_cents: int | None = Field(
        default=None, serialization_alias="maxPriceCents", validation_alias="maxPriceCents"
    )
    currency: str | None = None
    is_public: bool | None = Field(
        default=None, serialization_alias="isPublic", validation_alias="isPublic"
    )
    is_virtual: bool | None = Field(
        default=None, serialization_alias="isVirtual", validation_alias="isVirtual"
    )


class ListEventsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/events``."""

    data: list[Event]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/events``.

    All fields are optional; pass only what you need.
    """

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    status: EventStatus | None = None
    search: str | None = None
    start_before: str | None = Field(
        default=None, serialization_alias="startBefore", validation_alias="startBefore"
    )
    start_after: str | None = Field(
        default=None, serialization_alias="startAfter", validation_alias="startAfter"
    )
    sort_field: str | None = Field(
        default=None, serialization_alias="sortField", validation_alias="sortField"
    )
    sort_direction: Literal["asc", "desc"] | None = Field(
        default=None, serialization_alias="sortDirection", validation_alias="sortDirection"
    )
    filters: str | None = None
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/events/{id}``."""

    fields: str | None = None


class UpdateBody(_BaseModel):
    """Body shape accepted by ``PATCH /v1/events/{id}``.

    Only fields you provide are changed.
    """

    name: str | None = None
