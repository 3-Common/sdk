"""Contacts resource — sync and async clients plus public types.

Most callers reach this module through
[ThreeCommon.contacts][threecommon.ThreeCommon] /
[AsyncThreeCommon.contacts][threecommon.AsyncThreeCommon]; importing the
service classes directly is supported for advanced wiring.
"""

from threecommon.contacts.service import AsyncContactsService, ContactsService
from threecommon.contacts.types import (
    ActivityListParams,
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
    UpdateBody,
)

__all__ = (
    "ActivityListParams",
    "AsyncContactsService",
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
    "UpdateBody",
)
