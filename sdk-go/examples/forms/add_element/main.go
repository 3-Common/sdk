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

	required := true
	element, err := api.Forms.AddElement(context.Background(), "frm_123", &forms.AddElementParams{
		Type:     forms.ElementText,
		Prompt:   "What is your name?",
		Required: &required,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("added element %s (%s)\n", element.ID, element.Type)
}
