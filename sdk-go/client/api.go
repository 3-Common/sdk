// Package client is the recommended entry point for the SDK. [New] constructs
// a [*API] that holds one connection-pooled HTTP backend shared by every
// resource client.
//
//	api, err := client.New(threecommon.Config{APIKey: "..."})
//	if err != nil { log.Fatal(err) }
//
//	result, err := api.Events.List(ctx, &events.ListParams{Status: events.StatusOpen})
//
// Customers who only need a single resource can also instantiate that
// resource's client directly; the only difference is that
// each direct constructor builds its own backend rather than sharing.
package client

import (
	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
	"github.com/3-Common/sdk/sdk-go/resources/entitlements"
	"github.com/3-Common/sdk/sdk-go/resources/events"
	"github.com/3-Common/sdk/sdk-go/resources/features"
	"github.com/3-Common/sdk/sdk-go/resources/forms"
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
	"github.com/3-Common/sdk/sdk-go/resources/prices"
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

// API aggregates every resource the SDK exposes. Construct one with [New];
// the zero value is not usable.
type API struct {
	// Events is the events resource — GET /v1/events,
	// GET /v1/events/{id}, PATCH /v1/events/{id}.
	Events *events.Client

	// Invoices is the invoices resource — List, Retrieve, Create, Update,
	// Finalize, Void, RecordPayment.
	Invoices *invoices.Client

	// Subscriptions is the subscriptions resource — List, Retrieve, Create,
	// Update, Activate, Cancel, CancelImmediately, MarkUnpaid, Bill, Renew,
	// PreviewUpcomingInvoice.
	Subscriptions *subscriptions.Client

	// Contacts is the contacts resource — List, Count, Retrieve, Create,
	// Update, Delete, BulkUpsert, ListActivity, ListAutoPaginate,
	// ListActivityAutoPaginate.
	Contacts *contacts.Client

	// Entitlements is the entitlements resource — List, Retrieve, Lookup,
	// Grant, Consume, ListAutoPaginate.
	Entitlements *entitlements.Client

	// Prices is the prices resource — List, Retrieve, Create, Update,
	// Archive, Unarchive, ListAutoPaginate.
	Prices *prices.Client

	// Features is the features resource — List, Resolve, Retrieve, Create,
	// Update, Archive, Unarchive, ListAutoPaginate.
	Features *features.Client

	// Forms is the forms resource — List, Create, Retrieve, Update,
	// Duplicate, AddElement, UpdateElement, DeleteElement, MoveElement,
	// EnableOtherOption, DisableOtherOption, AddLogicRule, RemoveLogicRule,
	// ListAutoPaginate.
	Forms *forms.Client

	backend *core.Client
}

// New validates cfg, builds a single shared backend, and wires every
// resource client to it. Returns a [*threecommon.ValidationError] when cfg is
// missing required fields or has invalid values.
func New(cfg threecommon.Config) (*API, error) {
	backend, err := core.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &API{
		Events:        events.FromBackend(backend),
		Invoices:      invoices.FromBackend(backend),
		Subscriptions: subscriptions.FromBackend(backend),
		Contacts:      contacts.FromBackend(backend),
		Entitlements:  entitlements.FromBackend(backend),
		Prices:        prices.FromBackend(backend),
		Features:      features.FromBackend(backend),
		Forms:         forms.FromBackend(backend),
		backend:       backend,
	}, nil
}

// DisableTelemetry turns off opt-out client telemetry at runtime. The next
// request and all subsequent ones omit the Threecommon-Client-Telemetry
// header.
func (a *API) DisableTelemetry() {
	core.TelemetryFromClient(a.backend).Disable()
}
