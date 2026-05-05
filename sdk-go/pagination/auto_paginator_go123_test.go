//go:build go1.23

package pagination_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/3-Common/sdk/sdk-go/pagination"
)

func TestIter_All_YieldsValuesAndTerminalError(t *testing.T) {
	t.Parallel()

	pages := [][]string{{"a", "b"}, {"c"}}
	it := pagination.NewIter(0, func(page int) ([]string, bool, error) {
		if page >= len(pages) {
			return nil, false, nil
		}
		return pages[page], page < len(pages)-1, nil
	})

	var got []string
	for v, err := range it.All() {
		assert.NoError(t, err)
		got = append(got, v)
	}
	assert.Equal(t, []string{"a", "b", "c"}, got)
}

func TestIter_All_SurfacesPaginationError(t *testing.T) {
	t.Parallel()

	boom := errors.New("network")
	it := pagination.NewIter(0, func(page int) ([]int, bool, error) {
		if page == 0 {
			return []int{1, 2}, true, nil
		}
		return nil, false, boom
	})

	var values []int
	var seenErr error
	for v, err := range it.All() {
		if err != nil {
			seenErr = err
			break
		}
		values = append(values, v)
	}

	assert.Equal(t, []int{1, 2}, values)
	assert.ErrorIs(t, seenErr, boom)
}

func TestIter_All_HonorsBreak(t *testing.T) {
	t.Parallel()

	pages := [][]int{{1, 2, 3}, {4, 5}}
	calls := 0
	it := pagination.NewIter(0, func(page int) ([]int, bool, error) {
		calls++
		if page >= len(pages) {
			return nil, false, nil
		}
		return pages[page], page < len(pages)-1, nil
	})

	var got []int
	for v, err := range it.All() {
		assert.NoError(t, err)
		got = append(got, v)
		if v == 2 {
			break
		}
	}

	assert.Equal(t, []int{1, 2}, got)
	assert.Equal(t, 1, calls, "second page should not be fetched after break")
}
