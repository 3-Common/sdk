// Run with: go run ./examples/features/auto_paginate
//
// Iterates every active feature in the catalog, transparently fetching each
// page as the previous one drains.
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

	iter := api.Features.ListAutoPaginate(context.Background(), &features.ListParams{
		Active: threecommon.Bool(true),
	})

	count := 0
	quantity := 0
	for iter.Next() {
		count++
		if iter.Current().Type == features.TypeQuantity {
			quantity++
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("iterated %d active features\n", count)
	fmt.Printf("%d are quantity-typed\n", quantity)
}
