// Run with: go run ./examples/forms/add_logic_rule
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

	element, err := api.Forms.AddLogicRule(context.Background(), "frm_123", "elm_1", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_2",
		Condition: forms.LogicCondition{
			OptionIndices: []int{0},
			Operator:      forms.LogicOperatorAnyOf,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("added logic rule to element %s (%s)\n", element.ID, element.Type)
}
