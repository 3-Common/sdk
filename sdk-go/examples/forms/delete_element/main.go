// Run with: go run ./examples/forms/delete_element
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

	result, err := api.Forms.DeleteElement(context.Background(), "frm_123", "elm_123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("deleted element %s\n", result.DeletedElementID)
}
