// Run with: go run ./examples/forms/update
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

	form, err := api.Forms.Update(context.Background(), "frm_123", &forms.UpdateParams{
		Name:   threecommon.String("Renamed survey"),
		Status: forms.FormStatusActive,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated form %s -> %s (%s)\n", form.ID, form.Name, form.Status)
}
