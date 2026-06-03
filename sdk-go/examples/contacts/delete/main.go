// Run with: go run ./examples/contacts/delete
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

	result, err := api.Contacts.Delete(context.Background(), "cnt_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("deleted %s\n", result.ID)
}
