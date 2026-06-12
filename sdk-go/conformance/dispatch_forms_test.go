package conformance

// Forms-resource dispatcher for the conformance harness. Kept in its own file
// so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		return api.Forms.Create(ctx, decodeFormBody[forms.CreateParams](t, "create", sc.Call.Args))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		return api.Forms.Update(ctx, id, decodeFormBody[forms.UpdateParams](t, "update", sc.Call.Args))
	case "duplicate":
		// Duplicate is the one method whose body may be legitimately omitted:
		// an absent scenario body exercises the SDK's no-body path.
		id, _ := sc.Call.Args["id"].(string)
		var params *forms.DuplicateParams
		if _, ok := sc.Call.Args["body"]; ok {
			params = decodeFormBody[forms.DuplicateParams](t, "duplicate", sc.Call.Args)
		}
		return api.Forms.Duplicate(ctx, id, params)
	case "addElement":
		id, _ := sc.Call.Args["id"].(string)
		return api.Forms.AddElement(ctx, id, decodeFormBody[forms.AddElementParams](t, "addElement", sc.Call.Args))
	case "updateElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.UpdateElement(ctx, id, elementID, decodeFormBody[forms.UpdateElementParams](t, "updateElement", sc.Call.Args))
	case "deleteElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DeleteElement(ctx, id, elementID)
	case "moveElement":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.MoveElement(ctx, id, elementID, decodeFormBody[forms.MoveElementParams](t, "moveElement", sc.Call.Args))
	case "addLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.AddLogicRule(ctx, id, elementID, decodeFormBody[forms.AddLogicRuleParams](t, "addLogicRule", sc.Call.Args))
	case "removeLogicRule":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		targetElementID, _ := sc.Call.Args["targetElementId"].(string)
		return api.Forms.RemoveLogicRule(ctx, id, elementID, targetElementID)
	case "enableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.EnableOtherOption(ctx, id, elementID, decodeFormBody[forms.EnableOtherOptionParams](t, "enableOtherOption", sc.Call.Args))
	case "disableOtherOption":
		id, _ := sc.Call.Args["id"].(string)
		elementID, _ := sc.Call.Args["elementId"].(string)
		return api.Forms.DisableOtherOption(ctx, id, elementID)
	case "listAutoPaginate":
		iter := api.Forms.ListAutoPaginate(ctx, buildFormListParams(sc.Call.Args))
		var collected []forms.FormSummary
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported forms scenario method %q", sc.Call.Method)
	return nil, nil
}

// decodeFormBody round-trips the scenario's raw body through JSON into the
// typed params struct P. Unlike per-field copying, this forwards every field
// the SDK can express, so the dispatcher cannot silently truncate a scenario
// body. A missing body fails the scenario: every body-taking method except
// duplicate requires one, mirroring the Node and Python dispatchers.
func decodeFormBody[P any](t *testing.T, method string, args map[string]any) *P {
	t.Helper()

	body, ok := args["body"].(map[string]any)
	if !ok {
		t.Fatalf("forms.%s: scenario args.body is required", method)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("forms.%s: marshal scenario body: %v", method, err)
	}
	var p P
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("forms.%s: decode scenario body into %T: %v", method, p, err)
	}
	restoreExplicitNulls(&p, body)
	return &p
}

// restoreExplicitNulls re-applies JSON nulls that encoding/json drops while
// decoding: a null body value leaves a pointer field nil, indistinguishable
// from an absent key. For threecommon.Nullable fields, a scenario null must
// reach the wire as an explicit null (the API's "clear" signal), not be
// silently omitted.
func restoreExplicitNulls(params any, body map[string]any) {
	v := reflect.ValueOf(params).Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		raw, present := body[name]
		if !present || raw != nil {
			continue
		}
		f := v.Field(i)
		if f.Kind() != reflect.Pointer || !f.IsNil() {
			continue
		}
		elem := f.Type().Elem()
		if elem.Kind() != reflect.Struct {
			continue
		}
		isNull, ok := elem.FieldByName("IsNull")
		if !ok || isNull.Type.Kind() != reflect.Bool {
			continue
		}
		n := reflect.New(elem)
		n.Elem().FieldByName("IsNull").SetBool(true)
		f.Set(n)
	}
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

// roundTripFormBody decodes a scenario body into P and re-marshals it back to
// a generic map, so tests can assert nothing was dropped or retyped.
func roundTripFormBody[P any](t *testing.T, method string, body map[string]any) map[string]any {
	t.Helper()

	params := decodeFormBody[P](t, method, map[string]any{"body": body})
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

// TestFormScenarioBody_AddElementForwardsEveryField pins the dispatcher's body
// decoding against silent truncation: every key the addElement request schema
// defines must survive the round trip into AddElementParams unchanged.
func TestFormScenarioBody_AddElementForwardsEveryField(t *testing.T) {
	t.Parallel()

	// Values mirror what scenario JSON produces: numbers are float64.
	body := map[string]any{
		"type":         "Select One",
		"prompt":       "Pick one",
		"promptHidden": false,
		"helperText":   "helper",
		"placeholder":  "choose...",
		"required":     true,
		"propertyId":   "prop_1",
		"propertyData": map[string]any{"objectType": "contact"},
		"contactField": "status",
		"options":      []any{"A", "B"},
		"dropdown":     true,
		"otherPrompt":  "Other",
		"minChoices":   float64(1),
		"maxChoices":   float64(2),
		"min":          "2024-01-01",
		"max":          "2024-12-31",
		"accept":       "image/*",
		"logicGroups": []any{map[string]any{
			"revealedElementIndex": float64(1),
			"optionIndices":        []any{float64(0)},
			"operator":             "any_of",
		}},
		"content":    "hello",
		"imageUrl":   "https://x.test/a.png",
		"imageWidth": float64(300),
	}

	assert.Equal(t, body, roundTripFormBody[forms.AddElementParams](t, "addElement", body))
}

// TestFormScenarioBody_UpdateElementForwardsForEventItems pins the typed
// forEventItems union and date bounds through the updateElement body decode.
func TestFormScenarioBody_UpdateElementForwardsForEventItems(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"prompt": "Pick one",
		"min":    "2024-01-01",
		"max":    "2024-12-31",
		"forEventItems": []any{
			map[string]any{"type": "eventItem", "eventId": "evt_1", "itemId": "itm_1"},
			map[string]any{"type": "checkoutProduct", "checkoutId": "chk_1", "productId": "prd_1"},
		},
	}

	assert.Equal(t, body, roundTripFormBody[forms.UpdateElementParams](t, "updateElement", body))
}

// TestFormScenarioBody_PreservesExplicitNulls pins that a JSON null in a
// scenario body reaches the wire as an explicit null (the API's "clear"
// signal) instead of being silently dropped by pointer decoding.
func TestFormScenarioBody_PreservesExplicitNulls(t *testing.T) {
	t.Parallel()

	elementBody := map[string]any{
		"prompt":      "Pick one",
		"helperText":  nil,
		"logicGroups": nil,
	}
	assert.Equal(t, elementBody,
		roundTripFormBody[forms.UpdateElementParams](t, "updateElement", elementBody))

	formBody := map[string]any{
		"name":              "Registration",
		"nameHidden":        nil,
		"submitButtonAlign": nil,
	}
	assert.Equal(t, formBody, roundTripFormBody[forms.UpdateParams](t, "update", formBody))
}

// TestFormScenarioBody_AddLogicRuleForwardsBothConditionShapes pins both logic
// condition variants through the addLogicRule body decode.
func TestFormScenarioBody_AddLogicRuleForwardsBothConditionShapes(t *testing.T) {
	t.Parallel()

	selection := map[string]any{
		"revealedElementId": "elm_2",
		"condition": map[string]any{
			"optionIndices": []any{float64(0)},
			"operator":      "any_of",
		},
	}
	assert.Equal(t, selection, roundTripFormBody[forms.AddLogicRuleParams](t, "addLogicRule", selection))

	yesNo := map[string]any{
		"revealedElementId": "elm_2",
		"condition": map[string]any{
			"selectionType": "is",
			"value":         true,
		},
	}
	assert.Equal(t, yesNo, roundTripFormBody[forms.AddLogicRuleParams](t, "addLogicRule", yesNo))
}
