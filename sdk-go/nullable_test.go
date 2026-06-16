package threecommon_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

type nullableProbe struct {
	Note *threecommon.Nullable[string] `json:"note,omitempty"`
}

func TestNullable_MarshalStates(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   nullableProbe
		want string
	}{
		"absent": {nullableProbe{}, `{}`},
		"null":   {nullableProbe{Note: threecommon.Null[string]()}, `{"note":null}`},
		"value":  {nullableProbe{Note: threecommon.NullableOf("hi")}, `{"note":"hi"}`},
		"empty":  {nullableProbe{Note: threecommon.NullableOf("")}, `{"note":""}`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			raw, err := json.Marshal(tc.in)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(raw))
		})
	}
}

func TestNullable_Unmarshal(t *testing.T) {
	t.Parallel()

	var n threecommon.Nullable[string]
	require.NoError(t, json.Unmarshal([]byte(`"hi"`), &n))
	assert.Equal(t, "hi", n.Value)
	assert.False(t, n.IsNull)

	require.NoError(t, json.Unmarshal([]byte(`null`), &n))
	assert.True(t, n.IsNull)
}
