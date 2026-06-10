// Run with: go run ./examples/features/error_handling
//
// Demonstrate the typed error tree on the features surface. Each subtype wraps
// a *APIError; branch with errors.As.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	// A feature key is unique per host — recreating an existing key conflicts.
	_, err = api.Features.Create(context.Background(), &features.CreateParams{
		Key:  "api_calls",
		Name: "API calls",
		Type: features.TypeQuantity,
	})
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var (
		conflict   *threecommon.ConflictError
		validation *threecommon.ValidationError
		notFound   *threecommon.NotFoundError
		auth       *threecommon.AuthError
		rate       *threecommon.RateLimitError
		conn       *threecommon.ConnectionError
	)

	switch {
	case errors.As(err, &conflict):
		fmt.Println("a feature with this key already exists")
	case errors.As(err, &validation):
		fmt.Printf("validation: %s\n", validation.Message)
	case errors.As(err, &notFound):
		fmt.Println("feature not found")
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
