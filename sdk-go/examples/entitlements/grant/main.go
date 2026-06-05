// Run with: go run ./examples/entitlements/grant
//
// Manually grants entitlement units to a customer — admin top-ups, comp
// credits, or migration. Idempotent on GrantID: replaying the same id returns
// the existing record without double-crediting.
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

	ent, err := api.Entitlements.Grant(context.Background(), &entitlements.GrantParams{
		ContactID:  "cnt_replace_with_real_id",
		FeatureKey: "api_calls",
		Amount:     100,
		GrantID:    "grant_2026_q2_goodwill",
		Metadata:   map[string]string{"reason": "service-credit", "approvedBy": "ops"},
	})
	if err != nil {
		log.Fatal(err)
	}

	balance := int64(0)
	if ent.Balance != nil {
		balance = *ent.Balance
	}
	fmt.Printf("granted — %s now has %d %s\n", ent.ContactID, balance, ent.FeatureKey)
}
