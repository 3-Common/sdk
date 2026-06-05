// Run with: go run ./examples/entitlements/retrieve
//
// Retrieves a single entitlement record by ID, including its grant history.
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

	ent, err := api.Entitlements.Retrieve(context.Background(), "ent_replace_with_real_id", nil)
	if err != nil {
		log.Fatal(err)
	}

	balance, granted, consumed := int64(0), int64(0), int64(0)
	if ent.Balance != nil {
		balance = *ent.Balance
	}
	if ent.TotalGranted != nil {
		granted = *ent.TotalGranted
	}
	if ent.TotalConsumed != nil {
		consumed = *ent.TotalConsumed
	}

	fmt.Printf("entitlement %s [%s]\n", ent.ID, ent.FeatureKey)
	fmt.Printf("  contact        %s\n", ent.ContactID)
	fmt.Printf("  balance        %d\n", balance)
	fmt.Printf("  totalGranted   %d\n", granted)
	fmt.Printf("  totalConsumed  %d\n", consumed)
	for _, grant := range ent.Grants {
		fmt.Printf("  grant %s [%s] %d/%d remaining\n", grant.ID, grant.Source, grant.Remaining, grant.Amount)
	}
}
