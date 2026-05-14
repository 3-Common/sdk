package threecommon

// String returns a *string pointing at v. Convenient for populating optional
// pointer fields in request bodies:
//
//	api.Events.Update(ctx, "evt_123", &events.UpdateParams{
//		Name: threecommon.String("New name"),
//	})
func String(v string) *string { return &v }

// Int64 returns a *int64 pointing at v.
func Int64(v int64) *int64 { return &v }

// Int returns a *int pointing at v.
func Int(v int) *int { return &v }

// Bool returns a *bool pointing at v.
func Bool(v bool) *bool { return &v }

// Float64 returns a *float64 pointing at v.
func Float64(v float64) *float64 { return &v }
