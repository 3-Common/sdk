// Run with: go run ./examples/forms/update_element
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

	notRequired := false
	element, err := api.Forms.UpdateElement(context.Background(), "frm_123", "elm_1", &forms.UpdateElementParams{
		Prompt:   "What is your full name?",
		Required: &notRequired,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated element %s (%s)\n", element.ID, element.Prompt)
}
