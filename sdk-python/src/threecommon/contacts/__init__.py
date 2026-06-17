"""Contacts resource — sync and async clients plus public types.

Most callers reach this module through
[ThreeCommon.contacts][threecommon.ThreeCommon] /
[AsyncThreeCommon.contacts][threecommon.AsyncThreeCommon]; importing the
service classes directly is supported for advanced wiring.
"""

from threecommon.contacts.service import AsyncContactsService, ContactsService
from threecommon.contacts.types import (
    ActivityListParams,
    AttachPaymentMethodBody,
    AttachPaymentMethodResult,
    BulkUpsertBody,
    BulkUpsertItem,
    BulkUpsertResult,
    CompactContactStatus,
    Contact,
    ContactActivity,
    ContactActivityType,
    ContactMergeResolution,
    ContactProperty,
    ContactQuickFilter,
    ContactStatus,
    ContactUpdate,
    ContactWithOrderDetails,
    CountResult,
    CreateBody,
    DeleteResult,
    ListActivityResponse,
    ListContactsResponse,
    ListParams,
    PaymentMethod,
    PaymentMethodBillingDetails,
    PaymentMethodCard,
    PaymentMethodSetupIntent,
    PaymentMethodStatus,
    RemovedPaymentMethod,
    UpdateBody,
)

__all__ = (
    "ActivityListParams",
    "AsyncContactsService",
    "AttachPaymentMethodBody",
    "AttachPaymentMethodResult",
    "BulkUpsertBody",
    "BulkUpsertItem",
    "BulkUpsertResult",
    "CompactContactStatus",
    "Contact",
    "ContactActivity",
    "ContactActivityType",
    "ContactMergeResolution",
    "ContactProperty",
    "ContactQuickFilter",
    "ContactStatus",
    "ContactUpdate",
    "ContactWithOrderDetails",
    "ContactsService",
    "CountResult",
    "CreateBody",
    "DeleteResult",
    "ListActivityResponse",
    "ListContactsResponse",
    "ListParams",
    "PaymentMethod",
    "PaymentMethodBillingDetails",
    "PaymentMethodCard",
    "PaymentMethodSetupIntent",
    "PaymentMethodStatus",
    "RemovedPaymentMethod",
    "UpdateBody",
)
