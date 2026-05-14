package threecommon_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	threecommon "github.com/3-Common/sdk/sdk-go"
)

func TestHelpers_RoundTrip(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "x", *threecommon.String("x"))
	assert.Equal(t, int64(42), *threecommon.Int64(42))
	assert.Equal(t, 7, *threecommon.Int(7))
	assert.True(t, *threecommon.Bool(true))
	assert.InEpsilon(t, 1.5, *threecommon.Float64(1.5), 1e-9)
}

func TestHelpers_ReturnFreshPointers(t *testing.T) {
	t.Parallel()

	a := threecommon.String("same")
	b := threecommon.String("same")
	assert.NotSame(t, a, b)
}
