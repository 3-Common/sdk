// Run with: go run ./examples/events/auto_paginate
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	iter := api.Events.ListAutoPaginate(context.Background(), &events.ListParams{
		Status: events.StatusOpen,
	})

	count := 0
	for iter.Next() {
		ev := iter.Current()
		count++
		fmt.Printf("%4d. %s — %s\n", count, ev.ID, ev.Name)
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("walked %d events total\n", count)
}
