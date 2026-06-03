// Run with: go run ./examples/features/retrieve
//
// Retrieves a single feature by ID.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	feature, err := api.Features.Retrieve(context.Background(), "feat_replace_with_real_id", nil)
	if err != nil {
		log.Fatal(err)
	}

	active := false
	if feature.Active != nil {
		active = *feature.Active
	}
	fmt.Printf("feature %s [%s]\n", feature.ID, feature.Type)
	fmt.Printf("  key    %s\n", feature.Key)
	fmt.Printf("  name   %s\n", feature.Name)
	fmt.Printf("  active %v\n", active)
	if len(feature.EnumValues) > 0 {
		fmt.Printf("  values %v\n", feature.EnumValues)
	}
}
