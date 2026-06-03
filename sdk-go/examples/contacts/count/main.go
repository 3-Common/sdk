// Run with: go run ./examples/contacts/count
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	count, err := api.Contacts.Count(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("host has %d contacts\n", count)
}
