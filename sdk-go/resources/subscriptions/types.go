// Package subscriptions provides the subscriptions resource client for the
// 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Subscriptions]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	sub, err := api.Subscriptions.Create(ctx, &subscriptions.CreateParams{
//		ContactID: "cnt_42",
//		PriceID:   "price_7",
//		Quantity:  threecommon.Int64(1),
//		TrialDays: threecommon.Int(14),
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	subs, _ := subscriptions.New(threecommon.Config{APIKey: "..."})
//	result, err := subs.List(ctx, nil)
//
// Type names inside this package omit the "Subscription" prefix to avoid
// stutter (e.g. subscriptions.ListParams, not subscriptions.SubscriptionListParams).
package subscriptions

// Status is the lifecycle status of a subscription.
type Status string

// Status values returned by the API. Unknown values from a future API
// version will surface as the raw string.
const (
	StatusIncomplete Status = "incomplete"
	StatusTrialing   Status = "trialing"
	StatusActive     Status = "active"
	StatusPastDue    Status = "past_due"
	StatusCanceled   Status = "canceled"
	StatusUnpaid     Status = "unpaid"
)

// Item is one billed item on a subscription.
type Item struct {
	ID       string `json:"id"`
	PriceID  string `json:"priceId"`
	Quantity int64  `json:"quantity"`
}

// TaxID is a host tax-ID snapshot rolled forward onto each renewal invoice.
type TaxID struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Subscription is the resource shape returned by the API. Pointer fields and
// `omitempty` strings are populated only when the server returned them —
// list responses with a `Fields` filter omit unrequested values.
type Subscription struct {
	ID                 string            `json:"id"`
	HostID             string            `json:"hostId,omitempty"`
	ContactID          string            `json:"contactId,omitempty"`
	CustomerEmail      string            `json:"customerEmail,omitempty"`
	PriceID            string            `json:"priceId,omitempty"`
	Quantity           *int64            `json:"quantity,omitempty"`
	Items              []Item            `json:"items,omitempty"`
	Status             Status            `json:"status,omitempty"`
	CurrentPeriodStart string            `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   string            `json:"currentPeriodEnd,omitempty"`
	TrialStart         string            `json:"trialStart,omitempty"`
	TrialEnd           string            `json:"trialEnd,omitempty"`
	BillingCycleAnchor string            `json:"billingCycleAnchor,omitempty"`
	CancelAt           string            `json:"cancelAt,omitempty"`
	CancelAtPeriodEnd  *bool             `json:"cancelAtPeriodEnd,omitempty"`
	CanceledAt         string            `json:"canceledAt,omitempty"`
	CancelReason       string            `json:"cancelReason,omitempty"`
	EndedAt            string            `json:"endedAt,omitempty"`
	StartedAt          string            `json:"startedAt,omitempty"`
	DunningEnabled     *bool             `json:"dunningEnabled,omitempty"`
	FirstFailureAt     string            `json:"firstFailureAt,omitempty"`
	NextRetryAt        string            `json:"nextRetryAt,omitempty"`
	RetryCount         *int64            `json:"retryCount,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	TaxIDs             []TaxID           `json:"taxIds,omitempty"`
	AutoCharge         *bool             `json:"autoCharge,omitempty"`
	PaymentDueDays     *int64            `json:"paymentDueDays,omitempty"`
	TaxRate            *float64          `json:"taxRate,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	CreatedAt          string            `json:"createdAt,omitempty"`
	UpdatedAt          string            `json:"updatedAt,omitempty"`
}

// InvoiceRef is a slim invoice reference returned alongside renew/bill/update.
type InvoiceRef struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

// Proration is the proration summary returned by [Client.Update].
type Proration struct {
	NetAmountMinor int64 `json:"netAmountMinor"`
	DaysRemaining  int64 `json:"daysRemaining"`
	DaysInCycle    int64 `json:"daysInCycle"`
}

// UpdateResult is the value returned by [Client.Update].
type UpdateResult struct {
	Subscription Subscription `json:"subscription"`
	// Invoice is populated only when the proration produced a positive amount.
	Invoice   *InvoiceRef `json:"invoice,omitempty"`
	Proration Proration   `json:"proration"`
}

// BillResult is the value returned by [Client.Bill].
type BillResult struct {
	Subscription Subscription `json:"subscription"`
	Invoice      InvoiceRef   `json:"invoice"`
}

// RenewResult is the value returned by [Client.Renew].
type RenewResult struct {
	Subscription Subscription `json:"subscription"`
	// Invoice is populated only when the renewal advanced the period.
	Invoice *InvoiceRef `json:"invoice,omitempty"`
}

// InvoicePreviewLineItem is one line item on a subscription invoice preview.
type InvoicePreviewLineItem struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitAmount  int64  `json:"unitAmount"`
	ProductID   string `json:"productId,omitempty"`
	PriceID     string `json:"priceId,omitempty"`
}

// InvoicePreview is the non-persisted projection of the invoice the next
// renewal will generate (Stripe-style invoice.upcoming).
type InvoicePreview struct {
	CustomerID     string                   `json:"customerId"`
	SubscriptionID string                   `json:"subscriptionId"`
	Currency       string                   `json:"currency"`
	LineItems      []InvoicePreviewLineItem `json:"lineItems"`
	Subtotal       int64                    `json:"subtotal"`
	Total          int64                    `json:"total"`
	PeriodStart    string                   `json:"periodStart"`
	PeriodEnd      string                   `json:"periodEnd"`
}

// ListParams are the query parameters accepted by [Client.List] and [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default (20).
	PageSize *int

	// Status filters by lifecycle status. Empty includes all statuses.
	Status Status

	// ContactID filters by recipient contact id.
	ContactID string

	// PriceID filters by Price reference.
	PriceID string

	// Fields is a comma-separated list of fields to include in the response.
	// Empty returns all fields.
	Fields string
}

// RetrieveParams are the query parameters accepted by [Client.Retrieve].
type RetrieveParams struct {
	// Fields is a comma-separated list of fields to include in the response.
	Fields string
}

// CreateItem is one item on a multi-item subscription create body.
type CreateItem struct {
	PriceID  string `json:"priceId"`
	Quantity *int64 `json:"quantity,omitempty"`
}

// CreateParams is the body shape accepted by [Client.Create]. Use either
// (PriceID + Quantity) for the single-item shortcut, or Items for the
// multi-item path.
type CreateParams struct {
	PriceID            string            `json:"priceId,omitempty"`
	Quantity           *int64            `json:"quantity,omitempty"`
	Items              []CreateItem      `json:"items,omitempty"`
	ContactID          string            `json:"contactId,omitempty"`
	CustomerEmail      string            `json:"customerEmail,omitempty"`
	TrialDays          *int              `json:"trialDays,omitempty"`
	BillingCycleAnchor string            `json:"billingCycleAnchor,omitempty"`
	CancelAt           string            `json:"cancelAt,omitempty"`
	DunningEnabled     *bool             `json:"dunningEnabled,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	TaxIDs             []TaxID           `json:"taxIds,omitempty"`
	AutoCharge         *bool             `json:"autoCharge,omitempty"`
	PaymentDueDays     *int              `json:"paymentDueDays,omitempty"`
	TaxRate            *float64          `json:"taxRate,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Only fields
// with non-zero/non-nil values are sent.
type UpdateParams struct {
	PriceID        string   `json:"priceId,omitempty"`
	Quantity       *int64   `json:"quantity,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	TaxIDs         []TaxID  `json:"taxIds,omitempty"`
	TaxRate        *float64 `json:"taxRate,omitempty"`
	AutoCharge     *bool    `json:"autoCharge,omitempty"`
	DunningEnabled *bool    `json:"dunningEnabled,omitempty"`
	PaymentDueDays *int     `json:"paymentDueDays,omitempty"`
}

// CancelParams is the body shape accepted by [Client.Cancel]. May be nil to
// cancel without a reason.
type CancelParams struct {
	Reason string `json:"reason,omitempty"`
}

// CancelImmediatelyParams is the body shape accepted by [Client.CancelImmediately].
// May be nil.
type CancelImmediatelyParams struct {
	Reason string `json:"reason,omitempty"`
}

// ListResponse is the body returned by GET /v1/subscriptions.
type ListResponse struct {
	Data    []Subscription `json:"data"`
	HasMore bool           `json:"hasMore"`
}

// retrieveEnvelope is the {"data": Subscription} shape used by every
// detail-returning endpoint.
type retrieveEnvelope struct {
	Data Subscription `json:"data"`
}

// updateEnvelope is the response shape returned by PATCH /v1/subscriptions/{id}.
type updateEnvelope struct {
	Data      Subscription `json:"data"`
	Invoice   *InvoiceRef  `json:"invoice,omitempty"`
	Proration Proration    `json:"proration"`
}

// billEnvelope is the response shape returned by POST /v1/subscriptions/{id}/bill.
type billEnvelope struct {
	Data    Subscription `json:"data"`
	Invoice InvoiceRef   `json:"invoice"`
}

// renewEnvelope is the response shape returned by POST /v1/subscriptions/{id}/renew.
type renewEnvelope struct {
	Data    Subscription `json:"data"`
	Invoice *InvoiceRef  `json:"invoice,omitempty"`
}

// previewEnvelope is the response shape returned by GET /v1/subscriptions/{id}/upcoming.
type previewEnvelope struct {
	Data struct {
		Invoice *InvoicePreview `json:"invoice"`
	} `json:"data"`
}
