// Run with: go run ./examples/forms/duplicate
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

	dup, err := api.Forms.Duplicate(context.Background(), "frm_123", &forms.DuplicateParams{
		Name: "Customer survey (copy)",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("duplicated into %s (%s)\n", dup.ID, dup.Name)
}
