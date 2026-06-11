// Run with: go run ./examples/forms/enable_other_option
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

	element, err := api.Forms.EnableOtherOption(context.Background(), "frm_123", "elm_select", &forms.EnableOtherOptionParams{
		OtherPrompt: "Other (please specify)",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("enabled other option on %s -> %s\n", element.ID, element.Type)
}
