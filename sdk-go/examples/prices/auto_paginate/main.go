// Run with: go run ./examples/prices/auto_paginate
//
// Iterates every active price across all products, transparently fetching each
// page as the previous one drains.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/prices"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Prices.ListAutoPaginate(context.Background(), &prices.ListParams{
		Active: threecommon.Bool(true),
	})

	count := 0
	recurring := 0
	for iter.Next() {
		count++
		if iter.Current().Type == prices.TypeRecurring {
			recurring++
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("iterated %d active prices\n", count)
	fmt.Printf("%d are recurring\n", recurring)
}
