package threecommon

import "encoding/json"

// Nullable is an optional request field with three states: absent (a nil
// *Nullable, omitted from the request), an explicit JSON null (IsNull true,
// which the API reads as "clear this setting"), or a concrete Value.
//
// Construct values with [NullableOf] and explicit nulls with [Null]:
//
//	api.Forms.UpdateElement(ctx, "frm_123", "elm_1", &forms.UpdateElementParams{
//		Placeholder: threecommon.NullableOf("Pick one..."),
//		HelperText:  threecommon.Null[string](), // clear server-side
//	})
type Nullable[T any] struct {
	// Value is the concrete value to send. Ignored when IsNull is true.
	Value T
	// IsNull, when true, serializes the field as JSON null.
	IsNull bool
}

// NullableOf returns a [Nullable] carrying the concrete value v.
func NullableOf[T any](v T) *Nullable[T] { return &Nullable[T]{Value: v} }

// Null returns a [Nullable] that serializes as JSON null, which the API reads
// as "clear this setting".
func Null[T any]() *Nullable[T] { return &Nullable[T]{IsNull: true} }

// MarshalJSON implements [encoding/json.Marshaler].
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// UnmarshalJSON implements [encoding/json.Unmarshaler]. Note that
// encoding/json decodes a JSON null into a nil *Nullable struct field without
// calling this method, so the null branch applies only when decoding into a
// Nullable that already exists.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = Nullable[T]{IsNull: true}
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*n = Nullable[T]{Value: v}
	return nil
}
