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

	element, err := api.Forms.UpdateElement(context.Background(), "frm_123", "elm_1", &forms.UpdateElementParams{
		Prompt:   threecommon.String("What is your full name?"),
		Required: threecommon.NullableOf(false),
		// Null clears a nullable setting server-side (removes the helper text).
		HelperText: threecommon.Null[string](),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated element %s (%s)\n", element.ID, element.Prompt)
}
