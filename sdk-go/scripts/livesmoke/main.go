// Pre-release smoke test against the live API.
//
// Runs <= 12 calls and verifies the happy path + the common error paths
// across the events, invoices, and subscriptions resources. Used by
// .github/workflows/live-smoke.yml (maintainer-only).
//
// Required env:
//
//	THREECOMMON_API_KEY    — a real API key
//
// Optional env:
//
//	THREECOMMON_BASE_URL   — defaults to https://api.3common.com
//	SMOKE_EVENT_ID         — an event ID owned by the API-key host; if set,
//	                         exercises the events.Retrieve happy path
//	SMOKE_INVOICE_ID       — an invoice ID owned by the API-key host; if set,
//	                         exercises the invoices.Retrieve happy path
//	SMOKE_SUBSCRIPTION_ID  — a subscription ID owned by the API-key host; if
//	                         set, exercises the subscriptions.Retrieve happy path
//
// Run with: go run ./scripts/livesmoke
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

// missingObjectID is a syntactically valid 24-hex ObjectId that will not
// match any real record. The API rejects non-ObjectId strings with a 400
// before reaching the not-found path, so this must look well-formed to test
// 404s.
const missingObjectID = "000000000000000000000000"

type result struct {
	check  string
	status string // "pass", "fail", "skip"
	detail string
}

func main() {
	apiKey := os.Getenv("THREECOMMON_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "THREECOMMON_API_KEY env var is required for live-smoke runs")
		os.Exit(1)
	}

	baseURL := os.Getenv("THREECOMMON_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.3common.com"
	}
	knownEventID := os.Getenv("SMOKE_EVENT_ID")
	knownInvoiceID := os.Getenv("SMOKE_INVOICE_ID")
	knownSubscriptionID := os.Getenv("SMOKE_SUBSCRIPTION_ID")

	off := false
	api, err := client.New(threecommon.Config{
		APIKey:    apiKey,
		BaseURL:   baseURL,
		Telemetry: &off,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "client.New:", err)
		os.Exit(1)
	}

	results := []result{}
	ctx := context.Background()
	pageSize := 1

	// 1. List events.
	if r, listErr := api.Events.List(ctx, &events.ListParams{PageSize: &pageSize}); listErr == nil {
		results = append(results, result{
			check:  "events.List",
			status: "pass",
			detail: fmt.Sprintf("data.len=%d hasMore=%v", len(r.Data), r.HasMore),
		})
	} else {
		results = append(results, result{check: "events.List", status: "fail", detail: errMsg(listErr)})
	}

	// 2. Auto-paginate (one round of next()).
	iter := api.Events.ListAutoPaginate(ctx, &events.ListParams{PageSize: &pageSize})
	switch {
	case iter.Next():
		results = append(results, result{check: "events.ListAutoPaginate", status: "pass", detail: "yielded one"})
	case iter.Err() != nil:
		results = append(results, result{check: "events.ListAutoPaginate", status: "fail", detail: errMsg(iter.Err())})
	default:
		results = append(results, result{check: "events.ListAutoPaginate", status: "pass", detail: "empty"})
	}

	// 3. Retrieve a known event (if configured).
	if knownEventID != "" {
		if ev, retrieveErr := api.Events.Retrieve(ctx, knownEventID, nil); retrieveErr == nil {
			results = append(results, result{check: "events.Retrieve", status: "pass", detail: "id=" + ev.ID})
		} else {
			results = append(results, result{check: "events.Retrieve", status: "fail", detail: errMsg(retrieveErr)})
		}
	} else {
		results = append(results, result{check: "events.Retrieve", status: "skip", detail: "SMOKE_EVENT_ID not set"})
	}

	// 4. 404 path on events — well-formed ID that does not exist.
	if _, missErr := api.Events.Retrieve(ctx, missingObjectID, nil); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{
				check:  "events 404 path",
				status: "pass",
				detail: fmt.Sprintf("code=%s requestId=%s", nf.Code, nf.RequestID),
			})
		} else {
			results = append(results, result{check: "events 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "events 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	// 5. List invoices.
	if r, listErr := api.Invoices.List(ctx, &invoices.ListParams{PageSize: &pageSize}); listErr == nil {
		results = append(results, result{
			check:  "invoices.List",
			status: "pass",
			detail: fmt.Sprintf("data.len=%d hasMore=%v", len(r.Data), r.HasMore),
		})
	} else {
		results = append(results, result{check: "invoices.List", status: "fail", detail: errMsg(listErr)})
	}

	// 6. Retrieve a known invoice (if configured).
	if knownInvoiceID != "" {
		if inv, retrieveErr := api.Invoices.Retrieve(ctx, knownInvoiceID, nil); retrieveErr == nil {
			results = append(results, result{check: "invoices.Retrieve", status: "pass", detail: "id=" + inv.ID})
		} else {
			results = append(results, result{check: "invoices.Retrieve", status: "fail", detail: errMsg(retrieveErr)})
		}
	} else {
		results = append(results, result{check: "invoices.Retrieve", status: "skip", detail: "SMOKE_INVOICE_ID not set"})
	}

	// 7. 404 path on invoices.
	if _, missErr := api.Invoices.Retrieve(ctx, missingObjectID, nil); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{
				check:  "invoices 404 path",
				status: "pass",
				detail: fmt.Sprintf("code=%s requestId=%s", nf.Code, nf.RequestID),
			})
		} else {
			results = append(results, result{check: "invoices 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "invoices 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	// 7b. Not-found paths for the invoice write methods. The happy paths move
	// real money (AutoCharge, RefundPayment) or delete a record (DeleteDraft),
	// so only the 404 path is smoke-tested — a missing id 404s before any side
	// effect.
	if _, missErr := api.Invoices.AutoCharge(ctx, missingObjectID); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{check: "invoices.AutoCharge 404 path", status: "pass", detail: "code=" + nf.Code})
		} else {
			results = append(results, result{check: "invoices.AutoCharge 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "invoices.AutoCharge 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	if _, missErr := api.Invoices.RefundPayment(ctx, missingObjectID, missingObjectID, &invoices.RefundParams{Amount: 1}); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{check: "invoices.RefundPayment 404 path", status: "pass", detail: "code=" + nf.Code})
		} else {
			results = append(results, result{check: "invoices.RefundPayment 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "invoices.RefundPayment 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	if _, missErr := api.Invoices.DeleteDraft(ctx, missingObjectID); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{check: "invoices.DeleteDraft 404 path", status: "pass", detail: "code=" + nf.Code})
		} else {
			results = append(results, result{check: "invoices.DeleteDraft 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "invoices.DeleteDraft 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	// 8. List subscriptions.
	if r, listErr := api.Subscriptions.List(ctx, &subscriptions.ListParams{PageSize: &pageSize}); listErr == nil {
		results = append(results, result{
			check:  "subscriptions.List",
			status: "pass",
			detail: fmt.Sprintf("data.len=%d hasMore=%v", len(r.Data), r.HasMore),
		})
	} else {
		results = append(results, result{check: "subscriptions.List", status: "fail", detail: errMsg(listErr)})
	}

	// 9. Retrieve a known subscription (if configured).
	if knownSubscriptionID != "" {
		if sub, retrieveErr := api.Subscriptions.Retrieve(ctx, knownSubscriptionID, nil); retrieveErr == nil {
			results = append(results, result{check: "subscriptions.Retrieve", status: "pass", detail: "id=" + sub.ID})
		} else {
			results = append(results, result{check: "subscriptions.Retrieve", status: "fail", detail: errMsg(retrieveErr)})
		}
	} else {
		results = append(results, result{check: "subscriptions.Retrieve", status: "skip", detail: "SMOKE_SUBSCRIPTION_ID not set"})
	}

	// 10. 404 path on subscriptions.
	if _, missErr := api.Subscriptions.Retrieve(ctx, missingObjectID, nil); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{
				check:  "subscriptions 404 path",
				status: "pass",
				detail: fmt.Sprintf("code=%s requestId=%s", nf.Code, nf.RequestID),
			})
		} else {
			results = append(results, result{check: "subscriptions 404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "subscriptions 404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	// 11. 401 path — wrong API key.
	zero := 0
	//nolint:gosec // G101: deliberate fake to test the 401 path; not a real credential.
	const fakeKey = "3co_smoke_test_invalid_key" //gitleaks:allow
	bad, badErr := client.New(threecommon.Config{
		APIKey:     fakeKey,
		BaseURL:    baseURL,
		MaxRetries: &zero,
		Telemetry:  &off,
	})
	switch {
	case badErr != nil:
		results = append(results, result{check: "401 path", status: "fail", detail: errMsg(badErr)})
	default:
		_, callErr := bad.Events.List(ctx, &events.ListParams{PageSize: &pageSize})
		var ae *threecommon.AuthError
		switch {
		case errors.As(callErr, &ae):
			results = append(results, result{check: "401 path", status: "pass", detail: "code=" + ae.Code})
		case callErr == nil:
			results = append(results, result{check: "401 path", status: "fail", detail: "expected AuthError but call succeeded"})
		default:
			results = append(results, result{check: "401 path", status: "fail", detail: "unexpected error: " + errMsg(callErr)})
		}
	}

	failed := 0
	for _, r := range results {
		icon := "✓"
		switch r.status {
		case "fail":
			icon = "✗"
			failed++
		case "skip":
			icon = "○"
		}
		fmt.Printf("%s %s — %s\n", icon, r.check, r.detail)
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "\n%d check(s) failed.\n", failed)
		os.Exit(1)
	}
}

func errMsg(err error) string {
	var apiErr *threecommon.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Error()
	}
	return err.Error()
}
