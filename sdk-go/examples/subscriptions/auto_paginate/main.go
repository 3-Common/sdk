// Run with: go run ./examples/subscriptions/auto_paginate
//
// Iterates every active subscription, transparently fetching each page as the
// previous one drains.
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

	iter := api.Subscriptions.ListAutoPaginate(context.Background(), &subscriptions.ListParams{
		Status: subscriptions.StatusActive,
	})

	count := 0
	var units int64
	for iter.Next() {
		sub := iter.Current()
		count++
		if sub.Quantity != nil {
			units += *sub.Quantity
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("iterated %d active subscriptions\n", count)
	fmt.Printf("approximate units in flight: %d\n", units)
}
