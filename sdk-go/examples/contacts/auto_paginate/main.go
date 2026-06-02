// Run with: go run ./examples/contacts/auto_paginate
//
// Walk every opted-in contact for the host. Pages are fetched lazily —
// one HTTP call per page, only when the previous page's buffer drains.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Contacts.ListAutoPaginate(context.Background(), &contacts.ListParams{
		Filter: contacts.QuickFilterOptedIn,
	})

	total := 0
	lastEmail := ""
	for iter.Next() {
		c := iter.Current()
		total++
		lastEmail = c.Email
		if total%100 == 0 {
			fmt.Printf("...processed %d contacts\n", total)
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("walked %d opted-in contacts total (last: %s)\n", total, lastEmail)
}
