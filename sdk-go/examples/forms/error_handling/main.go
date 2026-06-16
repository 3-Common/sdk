// Run with: go run ./examples/forms/error_handling
//
// Demonstrate the typed error tree on the forms surface. Each subtype wraps
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
	"github.com/3-Common/sdk/sdk-go/resources/forms"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	_, err = api.Forms.Create(context.Background(), &forms.CreateParams{
		Name: "Registration",
		Type: forms.TypeStandalone,
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
		fmt.Println("form or element not found")
	case errors.As(err, &auth):
		fmt.Printf("auth failed: check your API key (code=%s)\n", auth.Code)
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
