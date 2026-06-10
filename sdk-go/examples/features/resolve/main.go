// Run with: go run ./examples/features/resolve
//
// Resolves a feature's live value for a customer — walks active subscriptions →
// prices → feature grants. For quantity features it also reports the current
// entitlement balance.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	resolved, err := api.Features.Resolve(context.Background(), &features.ResolveParams{
		ContactID:  "cnt_replace_with_real_id",
		FeatureKey: "api_calls",
	})
	if err != nil {
		log.Fatal(err)
	}

	v := resolved.Value
	fmt.Printf("feature %s [%s]\n", resolved.Feature.Key, v.Type)
	switch v.Type {
	case features.TypeBoolean:
		fmt.Printf("  enabled: %v\n", v.Enabled != nil && *v.Enabled)
	case features.TypeQuantity:
		if v.Quantity == nil {
			fmt.Println("  quantity: unlimited")
		} else {
			fmt.Printf("  quantity: %d\n", *v.Quantity)
		}
		if v.Balance != nil {
			fmt.Printf("  balance:  %d\n", *v.Balance)
		}
	case features.TypeEnum:
		val := "none"
		if v.EnumValue != nil {
			val = *v.EnumValue
		}
		fmt.Printf("  value: %s\n", val)
	case features.TypeDuration:
		if v.DurationDays == nil {
			fmt.Println("  duration: unlimited")
		} else {
			fmt.Printf("  duration: %d days\n", *v.DurationDays)
		}
	}
	fmt.Printf("  from subscriptions: %v\n", resolved.ContributingSubscriptionIDs)
}
