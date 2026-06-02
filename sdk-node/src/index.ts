// ── Public surface ─────────────────────────────────────────────────────────

// Main client.
export { ThreeCommon } from './client'

// Configuration.
export type { ClientConfig, Logger, RequestOptions } from './types/public'

// Events resource (types only — instances live on the client).
export type {
  Event,
  EventListParams,
  EventRetrieveParams,
  EventStatus,
  EventUpdateBody,
  EventsService,
  ListEventsResponse,
} from './resources/events'

// Invoices resource (types only — instances live on the client).
export type {
  AutoChargeOutcome,
  AutoChargeResult,
  DeletedInvoice,
  Invoice,
  InvoiceCreateBody,
  InvoiceCurrency,
  InvoiceLineItem,
  InvoiceListParams,
  InvoicePayment,
  InvoicePaymentBody,
  InvoiceRefundBody,
  InvoiceRetrieveParams,
  InvoiceStatus,
  InvoiceUpdateBody,
  InvoiceVoidBody,
  InvoicesService,
  ListInvoicesResponse,
} from './resources/invoices'

// Contacts resource (types only — instances live on the client).
export type {
  BulkUpsertContactsResult,
  Contact,
  ContactActivity,
  ContactActivityListParams,
  ContactActivityType,
  ContactBulkUpsertBody,
  ContactCountResult,
  ContactCreateBody,
  ContactListParams,
  ContactMergeResolution,
  ContactStatus,
  ContactUpdateBody,
  ContactWithOrderDetails,
  ContactsService,
  DeletedContact,
  ListContactActivityResponse,
  ListContactsResponse,
} from './resources/contacts'

// Subscriptions resource (types only — instances live on the client).
export type {
  BillSubscriptionResult,
  ListSubscriptionsResponse,
  RenewSubscriptionResult,
  Subscription,
  SubscriptionCancelBody,
  SubscriptionCancelImmediatelyBody,
  SubscriptionCreateBody,
  SubscriptionInvoicePreview,
  SubscriptionInvoicePreviewLineItem,
  SubscriptionInvoiceRef,
  SubscriptionItem,
  SubscriptionListParams,
  SubscriptionProration,
  SubscriptionRetrieveParams,
  SubscriptionStatus,
  SubscriptionTaxId,
  SubscriptionUpdateBody,
  SubscriptionsService,
  UpdateSubscriptionResult,
} from './resources/subscriptions'

// Filters — typed builder shared by every resource that accepts `filters`.
export { filter, and, combine, field, or } from './filters'
export type {
  FieldRef,
  FilterCondition,
  FilterGroup,
  FilterLogic,
  FilterOperator,
  FilterRange,
  FilterValue,
  Filters,
  SerializableFilter,
} from './filters'

// Errors. Every error thrown by the SDK is a subclass of ThreeCommonError.
export {
  ThreeCommonAuthError,
  ThreeCommonConflictError,
  ThreeCommonConnectionError,
  ThreeCommonError,
  ThreeCommonNotFoundError,
  ThreeCommonPermissionError,
  ThreeCommonRateLimitError,
  ThreeCommonServerError,
  ThreeCommonValidationError,
} from './errors'
export type { ErrorResponseBody } from './errors'

// Constants.
export { API_VERSION } from './api-version'
