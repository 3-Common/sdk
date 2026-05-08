// Run with: go run ./examples/events/error_handling
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

	_, err = api.Events.Retrieve(context.Background(), "000000000000000000000000", nil)
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var notFound *threecommon.NotFoundError
	var auth *threecommon.AuthError
	var rate *threecommon.RateLimitError
	var conn *threecommon.ConnectionError

	switch {
	case errors.As(err, &notFound):
		fmt.Printf("event not found — request_id=%s\n", notFound.RequestID)
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
