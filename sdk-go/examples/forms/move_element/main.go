// Run with: go run ./examples/forms/move_element
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

	form, err := api.Forms.MoveElement(context.Background(), "frm_123", "elm_1", &forms.MoveElementParams{
		Position: 2,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("moved element; form %s now has %d elements\n", form.ID, len(form.Elements))
}
