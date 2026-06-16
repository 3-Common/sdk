// Run with: go run ./examples/forms/create
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

	form, err := api.Forms.Create(context.Background(), &forms.CreateParams{
		Name:   "Registration",
		Type:   forms.TypeStandalone,
		Status: forms.StatusDraft,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created form %s (%s)\n", form.ID, form.Status)
}
