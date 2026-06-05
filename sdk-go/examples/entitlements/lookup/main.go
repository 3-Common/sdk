// Run with: go run ./examples/entitlements/lookup
//
// Looks up the unique entitlement for a (contact, feature) pair — the common
// "how much does this customer have left?" check. Returns a NotFoundError if
// no record exists yet.
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

	ent, err := api.Entitlements.Lookup(context.Background(), &entitlements.LookupParams{
		ContactID:  "cnt_replace_with_real_id",
		FeatureKey: "api_calls",
	})
	if err != nil {
		log.Fatal(err)
	}

	balance := int64(0)
	if ent.Balance != nil {
		balance = *ent.Balance
	}
	fmt.Printf("%s has %d %s remaining\n", ent.ContactID, balance, ent.FeatureKey)
}
