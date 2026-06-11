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
	case "listAutoPaginate":
		iter := api.Forms.ListAutoPaginate(ctx, buildFormListParams(sc.Call.Args))
		var collected []forms.FormSummary
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.Create(ctx, buildFormCreateParams(body))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		return api.Forms.Retrieve(ctx, id)
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
		return api.Forms.AddElement(ctx, id, buildAddElementParams(body))
	case "updateElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.UpdateElement(ctx, id, elementID, buildUpdateElementParams(body))
	case "deleteElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DeleteElement(ctx, id, elementID)
	case "moveElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.MoveElement(ctx, id, elementID, buildMoveElementParams(body))
	case "enableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.EnableOtherOption(ctx, id, elementID, buildEnableOtherOptionParams(body))
	case "disableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DisableOtherOption(ctx, id, elementID)
	case "addLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Forms.AddLogicRule(ctx, id, elementID, buildAddLogicRuleParams(body))
	case "removeLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		targetElementID, _ := sc.Call.Args["targetElementId"].(string)
		return api.Forms.RemoveLogicRule(ctx, id, elementID, targetElementID)
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
				p.Type = forms.FormType(s)
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
		p.Type = forms.FormType(s)
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.FormStatus(s)
	}
	if b, ok := body["nameHidden"].(bool); ok {
		p.NameHidden = &b
	}
	if s, ok := body["submitButtonText"].(string); ok {
		p.SubmitButtonText = s
	}
	return p
}

func buildFormUpdateParams(body map[string]any) *forms.UpdateParams {
	p := &forms.UpdateParams{}
	if body == nil {
		return p
	}
	if s, ok := body["name"].(string); ok {
		p.Name = &s
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.FormStatus(s)
	}
	if b, ok := body["nameHidden"].(bool); ok {
		p.NameHidden = &b
	}
	if s, ok := body["submitButtonText"].(string); ok {
		p.SubmitButtonText = &s
	}
	return p
}

func buildFormDuplicateParams(body map[string]any) *forms.DuplicateParams {
	p := &forms.DuplicateParams{}
	if body == nil {
		return p
	}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["status"].(string); ok {
		p.Status = forms.FormStatus(s)
	}
	return p
}

func buildAddElementParams(body map[string]any) *forms.AddElementParams {
	if body == nil {
		return nil
	}
	p := &forms.AddElementParams{}
	if s, ok := body["prompt"].(string); ok {
		p.Prompt = s
	}
	if s, ok := body["type"].(string); ok {
		p.Type = forms.ElementType(s)
	}
	if b, ok := body["required"].(bool); ok {
		p.Required = &b
	}
	if s, ok := body["otherPrompt"].(string); ok {
		p.OtherPrompt = s
	}
	return p
}

func buildUpdateElementParams(body map[string]any) *forms.UpdateElementParams {
	p := &forms.UpdateElementParams{}
	if body == nil {
		return p
	}
	if s, ok := body["prompt"].(string); ok {
		p.Prompt = &s
	}
	if b, ok := body["required"].(bool); ok {
		p.Required = &b
	}
	return p
}

func buildMoveElementParams(body map[string]any) *forms.MoveElementParams {
	if body == nil {
		return nil
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

func buildEnableOtherOptionParams(body map[string]any) *forms.EnableOtherOptionParams {
	if body == nil {
		return nil
	}
	p := &forms.EnableOtherOptionParams{}
	if s, ok := body["otherPrompt"].(string); ok {
		p.OtherPrompt = s
	}
	return p
}

func buildAddLogicRuleParams(body map[string]any) *forms.AddLogicRuleParams {
	if body == nil {
		return nil
	}
	p := &forms.AddLogicRuleParams{}
	if s, ok := body["revealedElementId"].(string); ok {
		p.RevealedElementID = s
	}
	if cond, ok := body["condition"].(map[string]any); ok {
		if s, ok := cond["operator"].(string); ok {
			p.Condition.Operator = forms.LogicOperator(s)
		}
		if s, ok := cond["selectionType"].(string); ok {
			p.Condition.SelectionType = forms.LogicSelectionType(s)
		}
		if b, ok := cond["value"].(bool); ok {
			p.Condition.Value = &b
		}
		if raw, ok := cond["optionIndices"].([]any); ok {
			for _, v := range raw {
				if n := anyToIntPtr(v); n != nil {
					p.Condition.OptionIndices = append(p.Condition.OptionIndices, *n)
				}
			}
		}
	}
	return p
}
