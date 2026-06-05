// Run with: go run ./examples/entitlements/consume
//
// Debits units from a customer's entitlement balance — call this when the
// customer uses the metered feature. Returns a ConflictError if the balance is
// insufficient.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/entitlements"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	ent, err := api.Entitlements.Consume(context.Background(), &entitlements.ConsumeParams{
		ContactID:  "cnt_replace_with_real_id",
		FeatureKey: "api_calls",
		Amount:     1,
		Reason:     "POST /v1/generate",
	})
	if err != nil {
		log.Fatal(err)
	}

	balance := int64(0)
	if ent.Balance != nil {
		balance = *ent.Balance
	}
	fmt.Printf("consumed 1 — %d %s remaining\n", balance, ent.FeatureKey)
}
