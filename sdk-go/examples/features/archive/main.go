// Run with: go run ./examples/features/archive
//
// Soft-archives a feature. Idempotent.
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

	feature, err := api.Features.Archive(context.Background(), "feat_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	active := false
	if feature.Active != nil {
		active = *feature.Active
	}
	fmt.Printf("archived %s — active=%v\n", feature.ID, active)
}
