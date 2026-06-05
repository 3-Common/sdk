// Run with: go run ./examples/entitlements/list
//
// Lists entitlement balance records, filtered by feature and a minimum balance.
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

	pageSize := 25
	result, err := api.Entitlements.List(context.Background(), &entitlements.ListParams{
		FeatureKey: "api_calls",
		MinBalance: threecommon.Int64(1),
		PageSize:   &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d entitlements (hasMore=%v)\n", len(result.Data), result.HasMore)
	for _, ent := range result.Data {
		balance := int64(0)
		if ent.Balance != nil {
			balance = *ent.Balance
		}
		fmt.Printf("%s — %s — %s — balance %d\n", ent.ID, ent.ContactID, ent.FeatureKey, balance)
	}
}
