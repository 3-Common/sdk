// Pre-release smoke test against the live API.
//
// Runs <= 10 calls and verifies the happy path + the four common error paths.
// Used by .github/workflows/live-smoke.yml (maintainer-only).
//
// Required env:
//
//	THREECOMMON_API_KEY   — a real API key
//
// Optional env:
//
//	THREECOMMON_BASE_URL  — defaults to https://api.3common.com
//	SMOKE_EVENT_ID        — an event ID known to belong to the API-key host;
//	                        required for retrieve / 403 / 422 checks
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
)

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

	// 1. List events.
	pageSize := 1
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

	// 4. 404 path — random ID that should not exist.
	if _, missErr := api.Events.Retrieve(ctx, "evt_smoke_test_nonexistent_xyz_999999", nil); missErr != nil {
		var nf *threecommon.NotFoundError
		if errors.As(missErr, &nf) {
			results = append(results, result{
				check:  "404 path",
				status: "pass",
				detail: fmt.Sprintf("code=%s requestId=%s", nf.Code, nf.RequestID),
			})
		} else {
			results = append(results, result{check: "404 path", status: "fail", detail: "unexpected error: " + errMsg(missErr)})
		}
	} else {
		results = append(results, result{check: "404 path", status: "fail", detail: "expected NotFoundError but call succeeded"})
	}

	// 5. 401 path — wrong API key.
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
