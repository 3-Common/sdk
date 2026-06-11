// Run with: go run ./examples/forms/list
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

	pageSize := 25
	result, err := api.Forms.List(context.Background(), &forms.ListParams{
		Type:     forms.FormTypeStandalone,
		PageSize: &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d forms (hasMore=%v)\n\n", len(result.Data), result.HasMore)
	for _, f := range result.Data {
		fmt.Printf("  %s - %s (%s, %d elements)\n", f.ID, f.Name, f.Status, f.NumElements)
	}
}
