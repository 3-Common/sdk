// Run with: go run ./examples/subscriptions/uncomp_next_cycle
//
// Removes a staged comp so the next renewal bills at full price again — the
// inverse of comp_next_cycle. A no-op when no comp is pending, and allowed on a
// subscription in any state.
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

	sub, err := api.Subscriptions.UncompNextCycle(context.Background(), "sub_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("subscription %s [%s]\n", sub.ID, sub.Status)
	fmt.Printf("  next renewal (%s) will bill at full price\n", sub.CurrentPeriodEnd)
}
