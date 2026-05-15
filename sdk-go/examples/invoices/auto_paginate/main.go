// Run with: go run ./examples/invoices/auto_paginate
//
// Iterates every open invoice for a customer and sums the amounts due.
// Pages are fetched lazily — one HTTP call per page, only when the previous
// page's buffer drains.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Invoices.ListAutoPaginate(context.Background(), &invoices.ListParams{
		Status:     invoices.StatusOpen,
		CustomerID: "cnt_replace_with_real_id",
	})

	var totalDue int64
	for iter.Next() {
		inv := iter.Current()
		if inv.AmountDue != nil {
			totalDue += *inv.AmountDue
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("total amount due across all open invoices: %d cents\n", totalDue)
}
