// Run with: go run ./examples/subscriptions/comp_next_cycle
//
// Stages a one-time fully-free (100% off) next renewal cycle. The next renewal
// consumes the comp exactly once, then billing resumes at full price. Rejected
// on a canceled or unpaid subscription.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	sub, err := api.Subscriptions.CompNextCycle(context.Background(), "sub_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("subscription %s [%s]\n", sub.ID, sub.Status)
	fmt.Printf("  next renewal (%s) will be comped\n", sub.CurrentPeriodEnd)
}
