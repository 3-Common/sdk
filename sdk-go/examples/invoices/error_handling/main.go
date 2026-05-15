// Run with: go run ./examples/invoices/error_handling
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Try to finalize a draft. The error tree below covers the realistic
	// failure modes: not-found, already-open (conflict), auth, rate limit,
	// and network errors.
	_, err = api.Invoices.Finalize(context.Background(), "000000000000000000000000")
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var notFound *threecommon.NotFoundError
	var conflict *threecommon.ConflictError
	var validation *threecommon.ValidationError
	var auth *threecommon.AuthError
	var rate *threecommon.RateLimitError
	var conn *threecommon.ConnectionError

	switch {
	case errors.As(err, &notFound):
		fmt.Printf("invoice not found — request_id=%s\n", notFound.RequestID)
	case errors.As(err, &conflict):
		// e.g. invoice_already_finalized — refetch + branch on Status.
		fmt.Printf("conflict: %s — request_id=%s\n", conflict.Code, conflict.RequestID)
	case errors.As(err, &validation):
		fmt.Printf("validation failed: %s — details=%+v\n", validation.Message, validation.Details)
	case errors.As(err, &auth):
		fmt.Printf("auth failed: check your API key — code=%s\n", auth.Code)
	case errors.As(err, &rate):
		wait := rate.RetryAfter
		if wait == 0 {
			wait = 30 * time.Second
		}
		fmt.Printf("rate limited; waiting %s before retry\n", wait)
	case errors.As(err, &conn):
		fmt.Printf("network error: %v\n", conn.Cause)
	default:
		fmt.Printf("unexpected error: %v\n", err)
	}
}
