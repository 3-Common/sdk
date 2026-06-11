// Run with: go run ./examples/forms/add_element
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

	element, err := api.Forms.AddElement(context.Background(), "frm_123", &forms.AddElementParams{
		Prompt:   "What is your name?",
		Type:     forms.ElementTypeText,
		Required: threecommon.Bool(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("added element %s (%s)\n", element.ID, element.Type)
}
