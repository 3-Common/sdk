// Run with: go run ./examples/properties/auto_paginate
//
// Walk every contact property for the host. Pages are fetched lazily - one
// HTTP call per page, only when the previous page's buffer drains.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/properties"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Properties.ListAutoPaginate(context.Background(), &properties.ListParams{
		ObjectType: properties.ObjectTypeContact,
	})

	total := 0
	for iter.Next() {
		p := iter.Current()
		total++
		fmt.Printf("  %s - %s (%s)\n", p.ID, p.Name, p.Type)
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("walked %d contact properties total\n", total)
}
