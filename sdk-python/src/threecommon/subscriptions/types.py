"""Public types for the subscriptions resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

#: Lifecycle status of a subscription. Future server-side values will arrive
#: as raw strings until the SDK is updated.
SubscriptionStatus = Literal[
    "incomplete",
    "trialing",
    "active",
    "past_due",
    "canceled",
    "unpaid",
]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class SubscriptionItem(_BaseModel):
    """One billed item on a subscription."""

    id: str
    price_id: str = Field(serialization_alias="priceId", validation_alias="priceId")
    quantity: int


class SubscriptionTaxId(_BaseModel):
    """Host tax-ID snapshot carried onto each renewal invoice."""

    type: str
    value: str


class Subscription(_BaseModel):
    """One subscription as returned by the API.

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
    customer_email: str | None = Field(
        default=None, serialization_alias="customerEmail", validation_alias="customerEmail"
    )
    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )
    quantity: int | None = None
    items: list[SubscriptionItem] | None = None
    status: SubscriptionStatus | None = None
    current_period_start: str | None = Field(
        default=None,
        serialization_alias="currentPeriodStart",
        validation_alias="currentPeriodStart",
    )
    current_period_end: str | None = Field(
        default=None,
        serialization_alias="currentPeriodEnd",
        validation_alias="currentPeriodEnd",
    )
    trial_start: str | None = Field(
        default=None, serialization_alias="trialStart", validation_alias="trialStart"
    )
    trial_end: str | None = Field(
        default=None, serialization_alias="trialEnd", validation_alias="trialEnd"
    )
    billing_cycle_anchor: str | None = Field(
        default=None,
        serialization_alias="billingCycleAnchor",
        validation_alias="billingCycleAnchor",
    )
    cancel_at: str | None = Field(
        default=None, serialization_alias="cancelAt", validation_alias="cancelAt"
    )
    cancel_at_period_end: bool | None = Field(
        default=None,
        serialization_alias="cancelAtPeriodEnd",
        validation_alias="cancelAtPeriodEnd",
    )
    canceled_at: str | None = Field(
        default=None, serialization_alias="canceledAt", validation_alias="canceledAt"
    )
    cancel_reason: str | None = Field(
        default=None, serialization_alias="cancelReason", validation_alias="cancelReason"
    )
    ended_at: str | None = Field(
        default=None, serialization_alias="endedAt", validation_alias="endedAt"
    )
    started_at: str | None = Field(
        default=None, serialization_alias="startedAt", validation_alias="startedAt"
    )
    dunning_enabled: bool | None = Field(
        default=None,
        serialization_alias="dunningEnabled",
        validation_alias="dunningEnabled",
    )
    first_failure_at: str | None = Field(
        default=None,
        serialization_alias="firstFailureAt",
        validation_alias="firstFailureAt",
    )
    next_retry_at: str | None = Field(
        default=None,
        serialization_alias="nextRetryAt",
        validation_alias="nextRetryAt",
    )
    retry_count: int | None = Field(
        default=None, serialization_alias="retryCount", validation_alias="retryCount"
    )
    notes: str | None = None
    tax_ids: list[SubscriptionTaxId] | None = Field(
        default=None, serialization_alias="taxIds", validation_alias="taxIds"
    )
    auto_charge: bool | None = Field(
        default=None, serialization_alias="autoCharge", validation_alias="autoCharge"
    )
    payment_due_days: int | None = Field(
        default=None,
        serialization_alias="paymentDueDays",
        validation_alias="paymentDueDays",
    )
    tax_rate: float | None = Field(
        default=None, serialization_alias="taxRate", validation_alias="taxRate"
    )
    metadata: dict[str, str] | None = None
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class SubscriptionInvoiceRef(_BaseModel):
    """Slim invoice reference returned alongside renew/bill/update responses."""

    id: str
    status: str
    total: int
    currency: str


class SubscriptionProration(_BaseModel):
    """Proration summary returned by ``PATCH /v1/subscriptions/{id}``."""

    net_amount_minor: int = Field(
        serialization_alias="netAmountMinor", validation_alias="netAmountMinor"
    )
    days_remaining: int = Field(
        serialization_alias="daysRemaining", validation_alias="daysRemaining"
    )
    days_in_cycle: int = Field(serialization_alias="daysInCycle", validation_alias="daysInCycle")


class SubscriptionInvoicePreviewLineItem(_BaseModel):
    """One line item on a subscription invoice preview."""

    description: str
    quantity: int
    unit_amount: int = Field(serialization_alias="unitAmount", validation_alias="unitAmount")
    product_id: str | None = Field(
        default=None, serialization_alias="productId", validation_alias="productId"
    )
    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )


class SubscriptionInvoicePreview(_BaseModel):
    """Non-persisted projection of the invoice the next renewal will generate."""

    customer_id: str = Field(serialization_alias="customerId", validation_alias="customerId")
    subscription_id: str = Field(
        serialization_alias="subscriptionId", validation_alias="subscriptionId"
    )
    currency: str
    line_items: list[SubscriptionInvoicePreviewLineItem] = Field(
        serialization_alias="lineItems", validation_alias="lineItems"
    )
    subtotal: int
    total: int
    period_start: str = Field(serialization_alias="periodStart", validation_alias="periodStart")
    period_end: str = Field(serialization_alias="periodEnd", validation_alias="periodEnd")


class ListSubscriptionsResponse(_BaseModel):
    """Successful response shape from ``GET /v1/subscriptions``."""

    data: list[Subscription]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class UpdateSubscriptionResult(_BaseModel):
    """Successful response shape from ``PATCH /v1/subscriptions/{id}``."""

    subscription: Subscription
    invoice: SubscriptionInvoiceRef | None = None
    proration: SubscriptionProration


class BillSubscriptionResult(_BaseModel):
    """Successful response shape from ``POST /v1/subscriptions/{id}/bill``."""

    subscription: Subscription
    invoice: SubscriptionInvoiceRef


class RenewSubscriptionResult(_BaseModel):
    """Successful response shape from ``POST /v1/subscriptions/{id}/renew``."""

    subscription: Subscription
    invoice: SubscriptionInvoiceRef | None = None


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/subscriptions``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    status: SubscriptionStatus | None = None
    contact_id: str | None = Field(
        default=None, serialization_alias="contactId", validation_alias="contactId"
    )
    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/subscriptions/{id}``."""

    fields: str | None = None


class CreateBodyItem(_BaseModel):
    """One item on a multi-item subscription create body."""

    price_id: str = Field(serialization_alias="priceId", validation_alias="priceId")
    quantity: int | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/subscriptions``."""

    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )
    quantity: int | None = None
    items: list[CreateBodyItem] | None = None
    contact_id: str | None = Field(
        default=None, serialization_alias="contactId", validation_alias="contactId"
    )
    customer_email: str | None = Field(
        default=None,
        serialization_alias="customerEmail",
        validation_alias="customerEmail",
    )
    trial_days: int | None = Field(
        default=None, serialization_alias="trialDays", validation_alias="trialDays"
    )
    billing_cycle_anchor: str | None = Field(
        default=None,
        serialization_alias="billingCycleAnchor",
        validation_alias="billingCycleAnchor",
    )
    cancel_at: str | None = Field(
        default=None, serialization_alias="cancelAt", validation_alias="cancelAt"
    )
    dunning_enabled: bool | None = Field(
        default=None,
        serialization_alias="dunningEnabled",
        validation_alias="dunningEnabled",
    )
    notes: str | None = None
    tax_ids: list[SubscriptionTaxId] | None = Field(
        default=None, serialization_alias="taxIds", validation_alias="taxIds"
    )
    auto_charge: bool | None = Field(
        default=None, serialization_alias="autoCharge", validation_alias="autoCharge"
    )
    payment_due_days: int | None = Field(
        default=None,
        serialization_alias="paymentDueDays",
        validation_alias="paymentDueDays",
    )
    tax_rate: float | None = Field(
        default=None, serialization_alias="taxRate", validation_alias="taxRate"
    )
    metadata: dict[str, str] | None = None


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/subscriptions/{id}``.

    Only fields you provide are changed.
    """

    price_id: str | None = Field(
        default=None, serialization_alias="priceId", validation_alias="priceId"
    )
    quantity: int | None = None
    notes: str | None = None
    tax_ids: list[SubscriptionTaxId] | None = Field(
        default=None, serialization_alias="taxIds", validation_alias="taxIds"
    )
    tax_rate: float | None = Field(
        default=None, serialization_alias="taxRate", validation_alias="taxRate"
    )
    auto_charge: bool | None = Field(
        default=None, serialization_alias="autoCharge", validation_alias="autoCharge"
    )
    dunning_enabled: bool | None = Field(
        default=None,
        serialization_alias="dunningEnabled",
        validation_alias="dunningEnabled",
    )
    payment_due_days: int | None = Field(
        default=None,
        serialization_alias="paymentDueDays",
        validation_alias="paymentDueDays",
    )


class CancelBody(_BaseModel):
    """Body accepted by ``POST /v1/subscriptions/{id}/cancel``."""

    reason: str | None = None


class CancelImmediatelyBody(_BaseModel):
    """Body accepted by ``POST /v1/subscriptions/{id}/cancel-immediately``."""

    reason: str | None = None
