// Run with: go run ./examples/entitlements/auto_paginate
//
// Iterates every entitlement for a feature, transparently fetching each page
// as the previous one drains. Handy for usage reports or sweeping for
// low-balance customers.
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

	iter := api.Entitlements.ListAutoPaginate(context.Background(), &entitlements.ListParams{
		FeatureKey: "api_calls",
	})

	count := 0
	low := 0
	for iter.Next() {
		ent := iter.Current()
		count++
		if ent.Balance != nil && *ent.Balance < 10 {
			low++
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("iterated %d entitlements\n", count)
	fmt.Printf("%d are running low (balance < 10)\n", low)
}
