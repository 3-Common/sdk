// Run with: go run ./examples/subscriptions/cancel
//
// Schedules cancellation at the end of the current period. The customer
// retains access until CurrentPeriodEnd; the next renewal transitions the
// subscription to canceled instead of advancing.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	sub, err := api.Subscriptions.Cancel(context.Background(), "sub_replace_with_real_id", &subscriptions.CancelParams{
		Reason: "Customer requested via support ticket #4821",
	})
	if err != nil {
		log.Fatal(err)
	}

	cancelAtPeriodEnd := false
	if sub.CancelAtPeriodEnd != nil {
		cancelAtPeriodEnd = *sub.CancelAtPeriodEnd
	}
	fmt.Printf("subscription %s [%s]\n", sub.ID, sub.Status)
	fmt.Printf("  cancelAtPeriodEnd: %v\n", cancelAtPeriodEnd)
	fmt.Printf("  access continues until %s\n", sub.CurrentPeriodEnd)
}
