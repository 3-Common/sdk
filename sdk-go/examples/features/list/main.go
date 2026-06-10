// Run with: go run ./examples/features/list
//
// Lists the feature catalog, filtered by value type and active status.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	pageSize := 25
	result, err := api.Features.List(context.Background(), &features.ListParams{
		Type:     features.TypeQuantity,
		Active:   threecommon.Bool(true),
		PageSize: &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d features (hasMore=%v)\n", len(result.Data), result.HasMore)
	for _, feature := range result.Data {
		fmt.Printf("%s — %s — %s\n", feature.ID, feature.Key, feature.Type)
	}
}
