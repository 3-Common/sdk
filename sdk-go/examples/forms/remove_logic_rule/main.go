// Run with: go run ./examples/forms/remove_logic_rule
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	element, err := api.Forms.RemoveLogicRule(context.Background(), "frm_123", "elm_select", "elm_followup")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("removed logic rule from %s (%d groups remain)\n", element.ID, len(element.LogicGroups))
}
