"""Public types for the entitlements resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

#: Source of an entitlement grant.
#:
#: * ``subscription_recurring`` — cycle grant from a subscription renewal.
#: * ``one_time_addon`` — top-up purchase (consumed first by ``consume``).
#: * ``manual`` — admin-applied grant.
EntitlementGrantSource = Literal[
    "subscription_recurring",
    "one_time_addon",
    "manual",
]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class EntitlementGrant(_BaseModel):
    """One grant in an entitlement's grant history."""

    id: str
    source: EntitlementGrantSource
    source_id: str | None = Field(
        default=None, serialization_alias="sourceId", validation_alias="sourceId"
    )
    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )
    amount: int
    remaining: int
    added_at: str = Field(serialization_alias="addedAt", validation_alias="addedAt")


class Entitlement(_BaseModel):
    """One entitlement balance record as returned by the API.

    Optional fields are populated only when the server returned them — list
    responses with a ``fields`` filter omit unrequested values.
    """

    id: str
    host_id: str | None = Field(
        default=None, serialization_alias="hostId", validation_alias="hostId"
    )
    contact_id: str | None = Field(
        default=None, serialization_alias="contactId", validation_alias="contactId"
    )
    feature_key: str | None = Field(
        default=None, serialization_alias="featureKey", validation_alias="featureKey"
    )
    balance: int | None = None
    grants: list[EntitlementGrant] | None = None
    total_granted: int | None = Field(
        default=None, serialization_alias="totalGranted", validation_alias="totalGranted"
    )
    total_consumed: int | None = Field(
        default=None, serialization_alias="totalConsumed", validation_alias="totalConsumed"
    )
    metadata: dict[str, str] | None = None
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class ListEntitlementsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/entitlements``."""

    data: list[Entitlement]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/entitlements``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    contact_id: str | None = Field(
        default=None, serialization_alias="contactId", validation_alias="contactId"
    )
    feature_key: str | None = Field(
        default=None, serialization_alias="featureKey", validation_alias="featureKey"
    )
    min_balance: int | None = Field(
        default=None, serialization_alias="minBalance", validation_alias="minBalance"
    )
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/entitlements/{id}``."""

    fields: str | None = None


class LookupParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/entitlements/lookup``."""

    contact_id: str = Field(serialization_alias="contactId", validation_alias="contactId")
    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    fields: str | None = None


class GrantBody(_BaseModel):
    """Body accepted by ``POST /v1/entitlements/grants``."""

    contact_id: str = Field(serialization_alias="contactId", validation_alias="contactId")
    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    amount: int
    grant_id: str = Field(serialization_alias="grantId", validation_alias="grantId")
    metadata: dict[str, str] | None = None


class ConsumeBody(_BaseModel):
    """Body accepted by ``POST /v1/entitlements/consume``."""

    contact_id: str = Field(serialization_alias="contactId", validation_alias="contactId")
    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    amount: int
    reason: str | None = None
