// Run with: go run ./examples/forms/disable_other_option
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

	element, err := api.Forms.DisableOtherOption(context.Background(), "frm_123", "elm_1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("disabled other option on element %s (%s)\n", element.ID, element.Type)
}
