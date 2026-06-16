// Run with: go run ./examples/properties/list
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/properties"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageSize := 50
	result, err := api.Properties.List(context.Background(), &properties.ListParams{
		ObjectType: properties.ObjectTypeContact,
		Status:     properties.StatusActive,
		PageSize:   &pageSize,
		Sort:       "name",
		Order:      "asc",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d properties (hasMore=%v)\n\n", len(result.Data), result.HasMore)
	for _, p := range result.Data {
		fmt.Printf("  %s - %s (%s, %s)\n", p.ID, p.Name, p.Type, p.ObjectType)
	}
}
