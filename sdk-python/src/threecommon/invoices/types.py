"""Public types for the invoices resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

#: Lifecycle status of an invoice. ``payment_failed`` is set when an off-session
#: auto-charge attempt is rejected (decline / SCA / no card); the invoice is
#: still owed and can be retried or paid manually.
InvoiceStatus = Literal["draft", "open", "payment_failed", "paid", "void"]

#: Invoice currency code; all line amounts must match.
InvoiceCurrency = Literal["USD", "CAD"]

#: Outcome of an auto-charge attempt.
AutoChargeOutcome = Literal["paid", "failed"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class InvoiceLineItem(_BaseModel):
    """One line on an invoice."""

    description: str
    quantity: int
    unit_amount: int = Field(serialization_alias="unitAmount", validation_alias="unitAmount")
    product_id: str | None = Field(
        default=None, serialization_alias="productId", validation_alias="productId"
    )
    tax_amount: int | None = Field(
        default=None, serialization_alias="taxAmount", validation_alias="taxAmount"
    )


class InvoicePayment(_BaseModel):
    """One recorded payment against an invoice."""

    id: str
    amount: int
    paid_at: str = Field(serialization_alias="paidAt", validation_alias="paidAt")
    idempotency_key: str | None = Field(
        default=None, serialization_alias="idempotencyKey", validation_alias="idempotencyKey"
    )
    note: str | None = None


class Invoice(_BaseModel):
    """One invoice as returned by the API.

    Optional fields are populated only when the server returned them — list
    responses with a ``fields`` filter omit unrequested values.
    """

    id: str
    host_id: str | None = Field(
        default=None, serialization_alias="hostId", validation_alias="hostId"
    )
    customer_id: str | None = Field(
        default=None, serialization_alias="customerId", validation_alias="customerId"
    )
    number: str | None = None
    currency: InvoiceCurrency | None = None
    line_items: list[InvoiceLineItem] | None = Field(
        default=None, serialization_alias="lineItems", validation_alias="lineItems"
    )
    payments: list[InvoicePayment] | None = None
    subtotal: int | None = None
    tax_total: int | None = Field(
        default=None, serialization_alias="taxTotal", validation_alias="taxTotal"
    )
    total: int | None = None
    amount_paid: int | None = Field(
        default=None, serialization_alias="amountPaid", validation_alias="amountPaid"
    )
    amount_due: int | None = Field(
        default=None, serialization_alias="amountDue", validation_alias="amountDue"
    )
    status: InvoiceStatus | None = None
    notes: str | None = None
    issued_at: str | None = Field(
        default=None, serialization_alias="issuedAt", validation_alias="issuedAt"
    )
    due_at: str | None = Field(default=None, serialization_alias="dueAt", validation_alias="dueAt")
    paid_at: str | None = Field(
        default=None, serialization_alias="paidAt", validation_alias="paidAt"
    )
    voided_at: str | None = Field(
        default=None, serialization_alias="voidedAt", validation_alias="voidedAt"
    )
    subscription_id: str | None = Field(
        default=None, serialization_alias="subscriptionId", validation_alias="subscriptionId"
    )
    quote_id: str | None = Field(
        default=None, serialization_alias="quoteId", validation_alias="quoteId"
    )
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class ListInvoicesResponse(_BaseModel):
    """Successful response shape from ``GET /v1/invoices``."""

    data: list[Invoice]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/invoices``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    status: InvoiceStatus | None = None
    customer_id: str | None = Field(
        default=None, serialization_alias="customerId", validation_alias="customerId"
    )
    subscription_id: str | None = Field(
        default=None, serialization_alias="subscriptionId", validation_alias="subscriptionId"
    )
    issued_after: str | None = Field(
        default=None, serialization_alias="issuedAfter", validation_alias="issuedAfter"
    )
    issued_before: str | None = Field(
        default=None, serialization_alias="issuedBefore", validation_alias="issuedBefore"
    )
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/invoices/{id}``."""

    fields: str | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/invoices``."""

    customer_id: str = Field(serialization_alias="customerId", validation_alias="customerId")
    currency: InvoiceCurrency
    line_items: list[InvoiceLineItem] = Field(
        serialization_alias="lineItems", validation_alias="lineItems"
    )
    notes: str | None = None
    due_at: str | None = Field(default=None, serialization_alias="dueAt", validation_alias="dueAt")
    subscription_id: str | None = Field(
        default=None, serialization_alias="subscriptionId", validation_alias="subscriptionId"
    )
    quote_id: str | None = Field(
        default=None, serialization_alias="quoteId", validation_alias="quoteId"
    )


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/invoices/{id}``.

    Only fields you provide are changed.
    """

    customer_id: str | None = Field(
        default=None, serialization_alias="customerId", validation_alias="customerId"
    )
    line_items: list[InvoiceLineItem] | None = Field(
        default=None, serialization_alias="lineItems", validation_alias="lineItems"
    )
    notes: str | None = None
    due_at: str | None = Field(default=None, serialization_alias="dueAt", validation_alias="dueAt")


class VoidBody(_BaseModel):
    """Body accepted by ``POST /v1/invoices/{id}/void``."""

    reason: str | None = None


class PaymentBody(_BaseModel):
    """Body accepted by ``POST /v1/invoices/{id}/payments``."""

    payment: int
    idempotency_key: str | None = Field(
        default=None, serialization_alias="idempotencyKey", validation_alias="idempotencyKey"
    )
    note: str | None = None


class RefundBody(_BaseModel):
    """Body accepted by ``POST /v1/invoices/{id}/payments/{paymentId}/refunds``."""

    amount: int
    reason: Literal["duplicate", "fraudulent", "requested_by_customer"] | None = None
    note: str | None = None
    idempotency_key: str | None = Field(
        default=None, serialization_alias="idempotencyKey", validation_alias="idempotencyKey"
    )


class AutoChargeResult(_BaseModel):
    """Successful response shape from ``POST /v1/invoices/{id}/auto_charge``.

    A card decline is an expected business outcome, not an error: ``outcome`` is
    ``"failed"`` with the invoice left in ``payment_failed`` and a
    ``failure_code`` set. Only network / processor 5xx errors raise.
    """

    invoice: Invoice
    outcome: AutoChargeOutcome
    failure_code: str | None = Field(
        default=None, serialization_alias="failureCode", validation_alias="failureCode"
    )


class DeletedInvoice(_BaseModel):
    """Result of ``DELETE /v1/invoices/{id}`` — the id of the removed draft."""

    id: str
