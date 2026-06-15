"""Public types for the contacts resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field

#: Lifecycle status of a contact.
#:
#: * ``opted-in`` / ``unsubscribed``: explicit consent state
#: * ``unknown``: never recorded a choice
#: * ``imported``: created via CSV / bulk-upsert before consent was captured
#: * ``deleted``: soft-deleted
ContactStatus = Literal["deleted", "imported", "unsubscribed", "opted-in", "unknown"]

#: Subset of statuses surfaced on the compact ``Contact`` projection
#: returned by ``list``, ``retrieve``, and ``create``.
CompactContactStatus = Literal["unsubscribed", "opted-in", "unknown"]

#: How to resolve field-level conflicts when merging a second contact into
#: the target during ``update``.
ContactMergeResolution = Literal["safe-merge", "overwrite-merge"]

#: The kind of event recorded against a contact in their activity feed.
ContactActivityType = Literal[
    "checkout_session_completed",
    "product_set_checkout_session_completed",
    "order_refunded",
    "ticket_scanned",
    "email_sent",
    "invoice_paid",
]

#: Quick status filter accepted by ``ListParams.filter``. Case-insensitive
#: on the wire; the SDK preserves casing.
ContactQuickFilter = Literal["all", "opted-in", "unknown", "unsubscribed", "imported"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class Contact(_BaseModel):
    """A contact in the compact projection returned by ``list``, ``retrieve``,
    and ``create``.

    Custom-property keys (24-char hex ids) may appear as additional top-level
    fields beyond those declared here — ``extra="ignore"`` on the base config
    means we silently drop them; access via ``model_extra`` if needed.
    """

    id: str
    first_name: str = Field(serialization_alias="firstName", validation_alias="firstName")
    last_name: str = Field(serialization_alias="lastName", validation_alias="lastName")
    full_name: str = Field(serialization_alias="fullName", validation_alias="fullName")
    email: str
    phone: str | None = None
    vendor_id: str = Field(serialization_alias="vendorId", validation_alias="vendorId")
    order_sum: float = Field(serialization_alias="orderSum", validation_alias="orderSum")
    gross_sum: float = Field(serialization_alias="grossSum", validation_alias="grossSum")
    first_order: int | None = Field(
        default=None, serialization_alias="firstOrder", validation_alias="firstOrder"
    )
    last_order: int | None = Field(
        default=None, serialization_alias="lastOrder", validation_alias="lastOrder"
    )
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    status: CompactContactStatus
    events_attended_ids: list[str] = Field(
        default_factory=list,
        serialization_alias="eventsAttended_IDS",
        validation_alias="eventsAttended_IDS",
    )
    items_purchased_ids: list[str] = Field(
        default_factory=list,
        serialization_alias="itemsPurchased_IDS",
        validation_alias="itemsPurchased_IDS",
    )
    products_purchased_ids: list[str] = Field(
        default_factory=list,
        serialization_alias="productsPurchased_IDS",
        validation_alias="productsPurchased_IDS",
    )


class ContactProperty(_BaseModel):
    """One custom-property entry on the richer order-details projection."""

    property_id: str = Field(serialization_alias="property_id", validation_alias="property_id")
    value: str | list[str] | bool


class ContactWithOrderDetails(_BaseModel):
    """The richer "order-details" projection returned by ``update``.

    Includes raw ``events_attended`` / ``items_purchased`` / ``products_purchased``
    arrays and the ``properties`` array, on top of everything in :class:`Contact`.
    The id field on this projection is ``_id`` (Mongo-style), not ``id``.
    """

    id_: str = Field(serialization_alias="_id", validation_alias="_id")
    email: str
    vendor_id: str = Field(serialization_alias="vendorId", validation_alias="vendorId")
    first_name: str = Field(serialization_alias="firstName", validation_alias="firstName")
    last_name: str = Field(serialization_alias="lastName", validation_alias="lastName")
    full_name: str = Field(serialization_alias="fullName", validation_alias="fullName")
    phone: str | None = None
    status: ContactStatus
    gross_sum: float = Field(serialization_alias="grossSum", validation_alias="grossSum")
    order_sum: float = Field(serialization_alias="orderSum", validation_alias="orderSum")
    least_recent_order: str | None = Field(
        default=None,
        serialization_alias="leastRecentOrder",
        validation_alias="leastRecentOrder",
    )
    most_recent_order: str | None = Field(
        default=None, serialization_alias="mostRecentOrder", validation_alias="mostRecentOrder"
    )
    events_attended: list[str] = Field(default_factory=list)
    items_purchased: list[str] = Field(default_factory=list)
    products_purchased: list[str] = Field(default_factory=list)
    properties: list[ContactProperty] | None = None
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class ContactActivity(_BaseModel):
    """A single activity record in a contact's activity feed."""

    id_: str = Field(serialization_alias="_id", validation_alias="_id")
    vendor_id: str = Field(serialization_alias="vendor_id", validation_alias="vendor_id")
    email: str
    contact_id: str | None = Field(
        default=None, serialization_alias="contact_id", validation_alias="contact_id"
    )
    type: ContactActivityType
    data: dict[str, Any]
    created_at: str = Field(serialization_alias="createdAt", validation_alias="createdAt")
    updated_at: str = Field(serialization_alias="updatedAt", validation_alias="updatedAt")


class ListContactsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/contacts``."""

    data: list[Contact]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")
    page_number: int = Field(serialization_alias="pageNumber", validation_alias="pageNumber")
    page_size: int = Field(serialization_alias="pageSize", validation_alias="pageSize")


class ListActivityResponse(_BaseModel):
    """Successful response shape from ``GET /v1/contacts/{id}/activity``."""

    data: list[ContactActivity]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")
    page_number: int = Field(serialization_alias="pageNumber", validation_alias="pageNumber")
    page_size: int = Field(serialization_alias="pageSize", validation_alias="pageSize")


class CountResult(_BaseModel):
    """Result shape returned by ``count``."""

    count: int


class BulkUpsertResult(_BaseModel):
    """Result shape returned by ``bulk_upsert``."""

    affected: int


class DeleteResult(_BaseModel):
    """Result shape returned by ``delete``. Echoes the removed contact id."""

    id: str


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/contacts``."""

    page_number: int | None = Field(
        default=None, serialization_alias="pageNumber", validation_alias="pageNumber"
    )
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    sort_field: str | None = Field(
        default=None, serialization_alias="sortField", validation_alias="sortField"
    )
    sort_direction: Literal["asc", "desc"] | None = Field(
        default=None, serialization_alias="sortDirection", validation_alias="sortDirection"
    )
    filter: ContactQuickFilter | None = None
    filters: str | None = None
    search: str | None = None


class ActivityListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/contacts/{id}/activity``."""

    page_number: int | None = Field(
        default=None, serialization_alias="pageNumber", validation_alias="pageNumber"
    )
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    filter: ContactActivityType | None = None
    sort: Literal["oldest"] | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/contacts``."""

    email: str
    first_name: str | None = Field(
        default=None, serialization_alias="firstName", validation_alias="firstName"
    )
    last_name: str | None = Field(
        default=None, serialization_alias="lastName", validation_alias="lastName"
    )
    phone: str | None = None


class ContactUpdate(_BaseModel):
    """The nested ``contact`` object inside :class:`UpdateBody`."""

    first_name: str = Field(serialization_alias="firstName", validation_alias="firstName")
    last_name: str = Field(serialization_alias="lastName", validation_alias="lastName")
    email: str
    phone: str | None = None
    status: ContactStatus


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/contacts/{id}``.

    The nested ``contact`` object carries the new field values; ``merge_with``
    and ``resolution`` are set together when an email change collides with
    another contact.
    """

    contact: ContactUpdate
    merge_with: str | None = Field(
        default=None, serialization_alias="mergeWith", validation_alias="mergeWith"
    )
    resolution: ContactMergeResolution | None = None


class BulkUpsertItem(_BaseModel):
    """One row in :class:`BulkUpsertBody.contacts`. Wider than :class:`CreateBody`
    to support CSV-import flows that carry status + properties + association
    arrays."""

    model_config = ConfigDict(
        populate_by_name=True,
        # Unlike other request bodies, the bulk endpoint accepts a `catchall`
        # of 24-char hex custom-property keys at the top level. Keep them.
        extra="allow",
        str_strip_whitespace=False,
    )

    email: str
    first_name: str | None = Field(
        default=None, serialization_alias="firstName", validation_alias="firstName"
    )
    last_name: str | None = Field(
        default=None, serialization_alias="lastName", validation_alias="lastName"
    )
    phone: str | None = None
    status: ContactStatus | None = None
    properties: list[ContactProperty] | None = None
    events_attended_ids: list[str] | None = Field(
        default=None,
        serialization_alias="eventsAttended_IDS",
        validation_alias="eventsAttended_IDS",
    )
    items_purchased_ids: list[str] | None = Field(
        default=None,
        serialization_alias="itemsPurchased_IDS",
        validation_alias="itemsPurchased_IDS",
    )
    products_purchased_ids: list[str] | None = Field(
        default=None,
        serialization_alias="productsPurchased_IDS",
        validation_alias="productsPurchased_IDS",
    )


class BulkUpsertBody(_BaseModel):
    """Body accepted by ``POST /v1/contacts/bulk``."""

    contacts: list[BulkUpsertItem]
