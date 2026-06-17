package conformance

// Subscriptions-resource dispatcher for the conformance harness. Kept in its
// own file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

func dispatchSubscriptions(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Subscriptions.List(ctx, buildSubscriptionListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *subscriptions.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &subscriptions.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Subscriptions.Retrieve(ctx, id, rp)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Subscriptions.Create(ctx, buildSubscriptionCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Subscriptions.Update(ctx, id, buildSubscriptionUpdateParams(body))
	case "retrieveManageUrl":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.RetrieveManageURL(ctx, id)
	case "activate":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.Activate(ctx, id)
	case "cancel":
		id, _ := sc.Call.Args["id"].(string)
		var cp *subscriptions.CancelParams
		if body, ok := sc.Call.Args["body"].(map[string]any); ok {
			cp = &subscriptions.CancelParams{}
			if r, ok := body["reason"].(string); ok {
				cp.Reason = r
			}
		}
		return api.Subscriptions.Cancel(ctx, id, cp)
	case "cancelImmediately":
		id, _ := sc.Call.Args["id"].(string)
		var cp *subscriptions.CancelImmediatelyParams
		if body, ok := sc.Call.Args["body"].(map[string]any); ok {
			cp = &subscriptions.CancelImmediatelyParams{}
			if r, ok := body["reason"].(string); ok {
				cp.Reason = r
			}
		}
		return api.Subscriptions.CancelImmediately(ctx, id, cp)
	case "markUnpaid":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.MarkUnpaid(ctx, id)
	case "bill":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.Bill(ctx, id)
	case "renew":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.Renew(ctx, id)
	case "previewUpcomingInvoice":
		id, _ := sc.Call.Args["id"].(string)
		return api.Subscriptions.PreviewUpcomingInvoice(ctx, id)
	case "listAutoPaginate":
		iter := api.Subscriptions.ListAutoPaginate(ctx, buildSubscriptionListParams(sc.Call.Args))
		var collected []subscriptions.Subscription
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported subscription scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildSubscriptionListParams(args map[string]any) *subscriptions.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &subscriptions.ListParams{}
	for k, v := range args {
		switch k {
		case "status":
			if s, ok := v.(string); ok {
				p.Status = subscriptions.Status(s)
			}
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "contactId":
			if s, ok := v.(string); ok {
				p.ContactID = s
			}
		case "priceId":
			if s, ok := v.(string); ok {
				p.PriceID = s
			}
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		}
	}
	return p
}

func buildSubscriptionTaxIDs(raw any) []subscriptions.TaxID {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]subscriptions.TaxID, 0, len(list))
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		t := subscriptions.TaxID{}
		if s, ok := m["type"].(string); ok {
			t.Type = s
		}
		if s, ok := m["value"].(string); ok {
			t.Value = s
		}
		out = append(out, t)
	}
	return out
}

func buildSubscriptionCreateItems(raw any) []subscriptions.CreateItem {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]subscriptions.CreateItem, 0, len(list))
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		it := subscriptions.CreateItem{}
		if s, ok := m["priceId"].(string); ok {
			it.PriceID = s
		}
		if q := anyToIntPtr(m["quantity"]); q != nil {
			n := int64(*q)
			it.Quantity = &n
		}
		out = append(out, it)
	}
	return out
}

func anyToInt64Ptr(v any) *int64 {
	if i := anyToIntPtr(v); i != nil {
		n := int64(*i)
		return &n
	}
	return nil
}

func anyToFloat64Ptr(v any) *float64 {
	switch n := v.(type) {
	case float64:
		return &n
	case float32:
		f := float64(n)
		return &f
	case int:
		f := float64(n)
		return &f
	case int64:
		f := float64(n)
		return &f
	}
	return nil
}

func anyToBoolPtr(v any) *bool {
	if b, ok := v.(bool); ok {
		return &b
	}
	return nil
}

func buildSubscriptionCreateParams(body map[string]any) *subscriptions.CreateParams {
	if body == nil {
		return nil
	}
	p := &subscriptions.CreateParams{}
	if s, ok := body["priceId"].(string); ok {
		p.PriceID = s
	}
	if q := anyToInt64Ptr(body["quantity"]); q != nil {
		p.Quantity = q
	}
	if raw, ok := body["items"]; ok {
		p.Items = buildSubscriptionCreateItems(raw)
	}
	if s, ok := body["contactId"].(string); ok {
		p.ContactID = s
	}
	if s, ok := body["customerEmail"].(string); ok {
		p.CustomerEmail = s
	}
	if td := anyToIntPtr(body["trialDays"]); td != nil {
		p.TrialDays = td
	}
	if s, ok := body["billingCycleAnchor"].(string); ok {
		p.BillingCycleAnchor = s
	}
	if s, ok := body["cancelAt"].(string); ok {
		p.CancelAt = s
	}
	if b := anyToBoolPtr(body["dunningEnabled"]); b != nil {
		p.DunningEnabled = b
	}
	if s, ok := body["notes"].(string); ok {
		p.Notes = s
	}
	if raw, ok := body["taxIds"]; ok {
		p.TaxIDs = buildSubscriptionTaxIDs(raw)
	}
	if b := anyToBoolPtr(body["autoCharge"]); b != nil {
		p.AutoCharge = b
	}
	if pd := anyToIntPtr(body["paymentDueDays"]); pd != nil {
		p.PaymentDueDays = pd
	}
	if tr := anyToFloat64Ptr(body["taxRate"]); tr != nil {
		p.TaxRate = tr
	}
	if md, ok := body["metadata"].(map[string]any); ok {
		out := map[string]string{}
		for k, v := range md {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
		p.Metadata = out
	}
	return p
}

func buildSubscriptionUpdateParams(body map[string]any) *subscriptions.UpdateParams {
	if body == nil {
		return &subscriptions.UpdateParams{}
	}
	p := &subscriptions.UpdateParams{}
	if s, ok := body["priceId"].(string); ok {
		p.PriceID = s
	}
	if q := anyToInt64Ptr(body["quantity"]); q != nil {
		p.Quantity = q
	}
	if s, ok := body["notes"].(string); ok {
		p.Notes = s
	}
	if raw, ok := body["taxIds"]; ok {
		p.TaxIDs = buildSubscriptionTaxIDs(raw)
	}
	if tr := anyToFloat64Ptr(body["taxRate"]); tr != nil {
		p.TaxRate = tr
	}
	if b := anyToBoolPtr(body["autoCharge"]); b != nil {
		p.AutoCharge = b
	}
	if b := anyToBoolPtr(body["dunningEnabled"]); b != nil {
		p.DunningEnabled = b
	}
	if pd := anyToIntPtr(body["paymentDueDays"]); pd != nil {
		p.PaymentDueDays = pd
	}
	return p
}
