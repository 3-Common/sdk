// Run with: go run ./examples/forms/add_logic_rule
//
// Conditional logic reveals a target element based on a source element's
// answer. Selection questions match on which options are chosen
// (OptionIndices + Operator); Yes/No questions match on the answer value
// (SelectionType + Value).
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

	// Selection question: reveal the target when the first option is chosen.
	element, err := api.Forms.AddLogicRule(context.Background(), "frm_123", "elm_select", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_2",
		Condition: forms.LogicCondition{
			OptionIndices: []int{0},
			Operator:      forms.LogicOperatorAnyOf,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("added selection rule to element %s (%s)\n", element.ID, element.Type)

	// Yes/No question: reveal the target when the respondent answers "yes".
	element, err = api.Forms.AddLogicRule(context.Background(), "frm_123", "elm_yes_no", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_3",
		Condition: forms.LogicCondition{
			SelectionType: forms.SelectionTypeIs,
			Value:         threecommon.Bool(true),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("added Yes/No rule to element %s (%s)\n", element.ID, element.Type)
}
