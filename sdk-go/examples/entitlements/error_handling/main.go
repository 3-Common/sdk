// Run with: go run ./examples/entitlements/error_handling
//
// Demonstrate the typed error tree on the entitlements surface. Each subtype
// wraps a *APIError; branch with errors.As.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/entitlements"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	_, err = api.Entitlements.Consume(context.Background(), &entitlements.ConsumeParams{
		ContactID:  "cnt_replace_with_real_id",
		FeatureKey: "api_calls",
		Amount:     1_000_000,
	})
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var (
		conflict   *threecommon.ConflictError
		notFound   *threecommon.NotFoundError
		validation *threecommon.ValidationError
		auth       *threecommon.AuthError
		rate       *threecommon.RateLimitError
		conn       *threecommon.ConnectionError
	)

	switch {
	case errors.As(err, &conflict):
		fmt.Println("insufficient balance — top up before consuming")
	case errors.As(err, &notFound):
		fmt.Printf("no entitlement record for this contact + feature — request_id=%s\n", notFound.RequestID)
	case errors.As(err, &validation):
		fmt.Printf("validation: %s\n", validation.Message)
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
