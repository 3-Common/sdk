package conformance

// Prices-resource dispatcher for the conformance harness. Kept in its own file
// so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/prices"
)

func dispatchPrices(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Prices.List(ctx, buildPriceListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *prices.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &prices.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Prices.Retrieve(ctx, id, rp)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Prices.Create(ctx, buildPriceCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Prices.Update(ctx, id, buildPriceUpdateParams(body))
	case "archive":
		id, _ := sc.Call.Args["id"].(string)
		return api.Prices.Archive(ctx, id)
	case "unarchive":
		id, _ := sc.Call.Args["id"].(string)
		return api.Prices.Unarchive(ctx, id)
	case "listAutoPaginate":
		iter := api.Prices.ListAutoPaginate(ctx, buildPriceListParams(sc.Call.Args))
		var collected []prices.Price
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported price scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildPriceListParams(args map[string]any) *prices.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &prices.ListParams{}
	for k, v := range args {
		switch k {
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "productId":
			if s, ok := v.(string); ok {
				p.ProductID = s
			}
		case "type":
			if s, ok := v.(string); ok {
				p.Type = prices.Type(s)
			}
		case "active":
			p.Active = anyToBoolPtr(v)
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		}
	}
	return p
}

func buildPriceRecurring(raw any) *prices.Recurring {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	r := &prices.Recurring{}
	if s, ok := m["interval"].(string); ok {
		r.Interval = prices.Interval(s)
	}
	if n := anyToInt64Ptr(m["intervalCount"]); n != nil {
		r.IntervalCount = *n
	}
	return r
}

func buildPriceFeatures(raw any) []prices.Feature {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]prices.Feature, 0, len(list))
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		f := prices.Feature{}
		if s, ok := m["featureKey"].(string); ok {
			f.FeatureKey = s
		}
		if s, ok := m["type"].(string); ok {
			f.Type = prices.FeatureType(s)
		}
		if b := anyToBoolPtr(m["enabled"]); b != nil {
			f.Enabled = b
		}
		if q, present := m["quantity"]; present {
			f.Quantity = anyToInt64Ptr(q)
		}
		if b := anyToBoolPtr(m["rolloverEnabled"]); b != nil {
			f.RolloverEnabled = b
		}
		if c := anyToInt64Ptr(m["rolloverCap"]); c != nil {
			f.RolloverCap = c
		}
		if b := anyToBoolPtr(m["expireOnCancel"]); b != nil {
			f.ExpireOnCancel = b
		}
		if s, ok := m["enumValue"].(string); ok {
			f.EnumValue = s
		}
		if d, present := m["durationDays"]; present {
			f.DurationDays = anyToInt64Ptr(d)
		}
		out = append(out, f)
	}
	return out
}

func priceMetadata(raw any) map[string]string {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func buildPriceCreateParams(body map[string]any) *prices.CreateParams {
	if body == nil {
		return nil
	}
	p := &prices.CreateParams{}
	if s, ok := body["productId"].(string); ok {
		p.ProductID = s
	}
	if s, ok := body["type"].(string); ok {
		p.Type = prices.Type(s)
	}
	if s, ok := body["currency"].(string); ok {
		p.Currency = prices.Currency(s)
	}
	if n := anyToInt64Ptr(body["unitAmount"]); n != nil {
		p.UnitAmount = *n
	}
	if raw, ok := body["recurring"]; ok {
		p.Recurring = buildPriceRecurring(raw)
	}
	if raw, ok := body["features"]; ok {
		p.Features = buildPriceFeatures(raw)
	}
	if s, ok := body["nickname"].(string); ok {
		p.Nickname = s
	}
	if raw, ok := body["metadata"]; ok {
		p.Metadata = priceMetadata(raw)
	}
	return p
}

func buildPriceUpdateParams(body map[string]any) *prices.UpdateParams {
	if body == nil {
		return &prices.UpdateParams{}
	}
	p := &prices.UpdateParams{}
	if n := anyToInt64Ptr(body["unitAmount"]); n != nil {
		p.UnitAmount = n
	}
	if raw, ok := body["recurring"]; ok {
		p.Recurring = buildPriceRecurring(raw)
	}
	if raw, ok := body["features"]; ok {
		p.Features = buildPriceFeatures(raw)
	}
	if s, ok := body["nickname"].(string); ok {
		p.Nickname = &s
	}
	if raw, ok := body["metadata"]; ok {
		p.Metadata = priceMetadata(raw)
	}
	return p
}
