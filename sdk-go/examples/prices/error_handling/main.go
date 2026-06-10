// Run with: go run ./examples/prices/error_handling
//
// Demonstrate the typed error tree on the prices surface. Each subtype wraps a
// *APIError; branch with errors.As.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/prices"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	// `recurring` is required when type is recurring; omitting it triggers a 400.
	_, err = api.Prices.Create(context.Background(), &prices.CreateParams{
		ProductID:  "prod_replace_with_real_id",
		Type:       prices.TypeRecurring,
		Currency:   prices.CurrencyUSD,
		UnitAmount: 1500,
	})
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var (
		validation *threecommon.ValidationError
		notFound   *threecommon.NotFoundError
		auth       *threecommon.AuthError
		rate       *threecommon.RateLimitError
		conn       *threecommon.ConnectionError
	)

	switch {
	case errors.As(err, &validation):
		fmt.Printf("validation: %s\n", validation.Message)
	case errors.As(err, &notFound):
		fmt.Println("product not found")
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
