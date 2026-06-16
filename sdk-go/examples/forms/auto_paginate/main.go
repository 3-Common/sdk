// Run with: go run ./examples/forms/auto_paginate
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/forms"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Forms.ListAutoPaginate(context.Background(), &forms.ListParams{
		Type: forms.TypeStandalone,
	})

	count := 0
	for iter.Next() {
		f := iter.Current()
		count++
		fmt.Printf("  %s - %s\n", f.ID, f.Name)
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nwalked %d forms\n", count)
}
