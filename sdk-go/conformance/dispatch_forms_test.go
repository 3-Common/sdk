package conformance

// Forms-resource dispatcher for the conformance harness. Kept in its own file
// so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/forms"
)

func dispatchForms(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Forms.List(ctx, buildFormListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		return api.Forms.Retrieve(ctx, id)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.Create(ctx, buildFormCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.Update(ctx, id, buildFormUpdateParams(body))
	case "duplicate":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.Duplicate(ctx, id, buildFormDuplicateParams(body))
	case "addElement":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.AddElement(ctx, id, buildFormAddElementParams(body))
	case "updateElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.UpdateElement(ctx, id, elementID, buildFormUpdateElementParams(body))
	case "deleteElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DeleteElement(ctx, id, elementID)
	case "moveElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.MoveElement(ctx, id, elementID, buildFormMoveElementParams(body))
	case "addLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.AddLogicRule(ctx, id, elementID, buildFormAddLogicRuleParams(body))
	case "removeLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		targetElementID, _ := sc.Call.Args["targetElementId"].(string)
		return api.Forms.RemoveLogicRule(ctx, id, elementID, targetElementID)
	case "enableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.EnableOtherOption(ctx, id, elementID, buildFormEnableOtherOptionParams(body))
	case "disableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DisableOtherOption(ctx, id, elementID)
	case "listAutoPaginate":
		iter := api.Forms.ListAutoPaginate(ctx, buildFormListParams(sc.Call.Args))
		var collected []forms.Form
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported forms scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildFormListParams(args map[string]any) *forms.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &forms.ListParams{}
	for k, v := range args {
		switch k {
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "type":
			if s, ok := v.(string); ok {
				p.Type = forms.Type(s)
			}
		}
	}
	return p
}

func buildFormCreateParams(body map[string]any) *forms.CreateParams {
	if body == nil {
		return nil
	}
	p := &forms.CreateParams{}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["type"].(string); ok {
		p.Type = forms.Type(s)
	}
	if b, ok := body["nameHidden"].(bool); ok {
		p.NameHidden = &b
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.Status(s)
	}
	if s, ok := body["submitButtonText"].(string); ok {
		p.SubmitButtonText = s
	}
	if s, ok := body["submitButtonWidth"].(string); ok {
		p.SubmitButtonWidth = forms.SubmitButtonWidth(s)
	}
	if s, ok := body["submitButtonAlign"].(string); ok {
		p.SubmitButtonAlign = forms.SubmitButtonAlign(s)
	}
	return p
}

func buildFormUpdateParams(body map[string]any) *forms.UpdateParams {
	if body == nil {
		return &forms.UpdateParams{}
	}
	p := &forms.UpdateParams{}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if b, ok := body["nameHidden"].(bool); ok {
		p.NameHidden = &b
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.Status(s)
	}
	if s, ok := body["submitButtonText"].(string); ok {
		p.SubmitButtonText = s
	}
	if s, ok := body["submitButtonWidth"].(string); ok {
		p.SubmitButtonWidth = forms.SubmitButtonWidth(s)
	}
	if s, ok := body["submitButtonAlign"].(string); ok {
		p.SubmitButtonAlign = forms.SubmitButtonAlign(s)
	}
	return p
}

func buildFormDuplicateParams(body map[string]any) *forms.DuplicateParams {
	if body == nil {
		return nil
	}
	p := &forms.DuplicateParams{}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.Status(s)
	}
	return p
}

func buildFormAddElementParams(body map[string]any) *forms.AddElementParams {
	if body == nil {
		return nil
	}
	p := &forms.AddElementParams{}
	if s, ok := body["type"].(string); ok {
		p.Type = forms.ElementType(s)
	}
	if s, ok := body["prompt"].(string); ok {
		p.Prompt = s
	}
	if b, ok := body["required"].(bool); ok {
		p.Required = &b
	}
	if raw, ok := body["options"].([]any); ok {
		p.Options = anyToStringSlice(raw)
	}
	return p
}

func buildFormUpdateElementParams(body map[string]any) *forms.UpdateElementParams {
	if body == nil {
		return &forms.UpdateElementParams{}
	}
	p := &forms.UpdateElementParams{}
	if s, ok := body["prompt"].(string); ok {
		p.Prompt = s
	}
	if b, ok := body["required"].(bool); ok {
		p.Required = &b
	}
	if raw, ok := body["options"].([]any); ok {
		p.Options = anyToStringSlice(raw)
	}
	return p
}

func buildFormMoveElementParams(body map[string]any) *forms.MoveElementParams {
	if body == nil {
		return &forms.MoveElementParams{}
	}
	p := &forms.MoveElementParams{}
	if n := anyToIntPtr(body["position"]); n != nil {
		p.Position = *n
	}
	if s, ok := body["section"].(string); ok {
		p.Section = forms.MoveSection(s)
	}
	return p
}

func buildFormAddLogicRuleParams(body map[string]any) *forms.AddLogicRuleParams {
	if body == nil {
		return nil
	}
	p := &forms.AddLogicRuleParams{}
	if s, ok := body["revealedElementId"].(string); ok {
		p.RevealedElementID = s
	}
	if cond, ok := body["condition"].(map[string]any); ok {
		if raw, ok := cond["optionIndices"].([]any); ok {
			p.Condition.OptionIndices = anyToIntSlice(raw)
		}
		if s, ok := cond["operator"].(string); ok {
			p.Condition.Operator = forms.LogicOperator(s)
		}
		if s, ok := cond["selectionType"].(string); ok {
			p.Condition.SelectionType = forms.SelectionType(s)
		}
		if b, ok := cond["value"].(bool); ok {
			p.Condition.Value = &b
		}
	}
	return p
}

func buildFormEnableOtherOptionParams(body map[string]any) *forms.EnableOtherOptionParams {
	if body == nil {
		return nil
	}
	p := &forms.EnableOtherOptionParams{}
	if s, ok := body["otherPrompt"].(string); ok {
		p.OtherPrompt = s
	}
	return p
}

func anyToStringSlice(raw []any) []string {
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func anyToIntSlice(raw []any) []int {
	out := make([]int, 0, len(raw))
	for _, v := range raw {
		if n := anyToIntPtr(v); n != nil {
			out = append(out, *n)
		}
	}
	return out
}
