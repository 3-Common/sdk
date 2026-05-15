// Package invoices provides the invoices resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Invoices]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	inv, err := api.Invoices.Create(ctx, &invoices.CreateParams{
//		CustomerID: "cnt_42",
//		Currency:   invoices.CurrencyUSD,
//		LineItems: []invoices.LineItem{
//			{Description: "Consulting", Quantity: 1, UnitAmount: 50_000},
//		},
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	inv, _ := invoices.New(threecommon.Config{APIKey: "..."})
//	result, err := inv.List(ctx, nil)
//
// Type names inside this package omit the "Invoice" prefix to avoid stutter
// (e.g. invoices.ListParams, not invoices.InvoiceListParams).
package invoices

// Status is the lifecycle status of an invoice.
type Status string

// Status values returned by the API. Unknown values from a future API
// version will surface as the raw string.
const (
	StatusDraft Status = "draft"
	StatusOpen  Status = "open"
	StatusPaid  Status = "paid"
	StatusVoid  Status = "void"
)

// Currency is the currency code for an invoice; all line amounts must match.
type Currency string

// Currency values supported by the API.
const (
	CurrencyUSD Currency = "USD"
	CurrencyCAD Currency = "CAD"
)

// LineItem is one line on an invoice.
type LineItem struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitAmount  int64  `json:"unitAmount"`
	ProductID   string `json:"productId,omitempty"`
	TaxAmount   *int64 `json:"taxAmount,omitempty"`
}

// Payment is one recorded payment against an invoice.
type Payment struct {
	ID             string `json:"id"`
	Amount         int64  `json:"amount"`
	PaidAt         string `json:"paidAt"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
	Note           string `json:"note,omitempty"`
}

// Invoice is the resource shape returned by the API. Pointer fields and
// `omitempty` strings are populated only when the server returned them —
// list responses with a `Fields` filter omit unrequested values.
type Invoice struct {
	ID             string     `json:"id"`
	HostID         string     `json:"hostId,omitempty"`
	CustomerID     string     `json:"customerId,omitempty"`
	Number         *string    `json:"number,omitempty"`
	Currency       Currency   `json:"currency,omitempty"`
	LineItems      []LineItem `json:"lineItems,omitempty"`
	Payments       []Payment  `json:"payments,omitempty"`
	Subtotal       *int64     `json:"subtotal,omitempty"`
	TaxTotal       *int64     `json:"taxTotal,omitempty"`
	Total          *int64     `json:"total,omitempty"`
	AmountPaid     *int64     `json:"amountPaid,omitempty"`
	AmountDue      *int64     `json:"amountDue,omitempty"`
	Status         Status     `json:"status,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	IssuedAt       string     `json:"issuedAt,omitempty"`
	DueAt          string     `json:"dueAt,omitempty"`
	PaidAt         string     `json:"paidAt,omitempty"`
	VoidedAt       string     `json:"voidedAt,omitempty"`
	SubscriptionID string     `json:"subscriptionId,omitempty"`
	QuoteID        string     `json:"quoteId,omitempty"`
	CreatedAt      string     `json:"createdAt,omitempty"`
	UpdatedAt      string     `json:"updatedAt,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default (20).
	PageSize *int

	// Status filters by lifecycle status. Empty includes all statuses.
	Status Status

	// CustomerID filters by recipient contact id.
	CustomerID string

	// IssuedAfter is an ISO 8601 timestamp; only invoices issued on or after this date are returned.
	IssuedAfter string

	// IssuedBefore is an ISO 8601 timestamp; only invoices issued on or before this date are returned.
	IssuedBefore string

	// Fields is a comma-separated list of fields to include in the response.
	// Empty returns all fields.
	Fields string
}

// RetrieveParams are the query parameters accepted by [Client.Retrieve].
type RetrieveParams struct {
	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// CreateParams is the body shape accepted by [Client.Create].
type CreateParams struct {
	CustomerID     string     `json:"customerId"`
	Currency       Currency   `json:"currency"`
	LineItems      []LineItem `json:"lineItems"`
	Notes          string     `json:"notes,omitempty"`
	DueAt          string     `json:"dueAt,omitempty"`
	SubscriptionID string     `json:"subscriptionId,omitempty"`
	QuoteID        string     `json:"quoteId,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Only fields
// with non-zero values are sent.
type UpdateParams struct {
	CustomerID string     `json:"customerId,omitempty"`
	LineItems  []LineItem `json:"lineItems,omitempty"`
	Notes      string     `json:"notes,omitempty"`
	DueAt      string     `json:"dueAt,omitempty"`
}

// VoidParams is the body shape accepted by [Client.Void]. May be nil to void
// without a reason.
type VoidParams struct {
	Reason string `json:"reason,omitempty"`
}

// PaymentParams is the body shape accepted by [Client.RecordPayment].
type PaymentParams struct {
	Payment        int64  `json:"payment"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
	Note           string `json:"note,omitempty"`
}

// ListResponse is the body returned by GET /v1/invoices.
type ListResponse struct {
	Data    []Invoice `json:"data"`
	HasMore bool      `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Invoice} shape used by every detail-returning endpoint.
type retrieveEnvelope struct {
	Data Invoice `json:"data"`
}
