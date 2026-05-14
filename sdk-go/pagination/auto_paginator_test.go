package pagination_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/3-Common/sdk/sdk-go/pagination"
)

func TestIter_WalksMultiplePages(t *testing.T) {
	t.Parallel()

	pages := [][]int{{1, 2}, {3}}
	calls := 0
	it := pagination.NewIter(0, func(page int) ([]int, bool, error) {
		calls++
		if page >= len(pages) {
			return nil, false, nil
		}
		return pages[page], page < len(pages)-1, nil
	})

	var out []int
	for it.Next() {
		out = append(out, it.Current())
	}

	require := assert.New(t)
	require.NoError(it.Err())
	require.Equal([]int{1, 2, 3}, out)
	require.Equal(2, calls)
}

func TestIter_StopsOnEmptyPage(t *testing.T) {
	t.Parallel()

	it := pagination.NewIter(0, func(_ int) ([]int, bool, error) {
		return nil, false, nil
	})

	assert.False(t, it.Next())
	assert.NoError(t, it.Err())
	assert.Zero(t, it.Current())
}

func TestIter_ReportsPageError(t *testing.T) {
	t.Parallel()

	boom := errors.New("network")
	it := pagination.NewIter(0, func(page int) ([]int, bool, error) {
		if page == 0 {
			return []int{1}, true, nil
		}
		return nil, false, boom
	})

	var out []int
	for it.Next() {
		out = append(out, it.Current())
	}

	assert.Equal(t, []int{1}, out)
	assert.ErrorIs(t, it.Err(), boom)

	// Subsequent Next calls remain false and Err is sticky.
	assert.False(t, it.Next())
	assert.ErrorIs(t, it.Err(), boom)
}

func TestIter_StartPageIsRespected(t *testing.T) {
	t.Parallel()

	calls := []int{}
	it := pagination.NewIter(3, func(page int) ([]int, bool, error) {
		calls = append(calls, page)
		return []int{page * 10}, false, nil
	})

	assert.True(t, it.Next())
	assert.Equal(t, 30, it.Current())
	assert.False(t, it.Next())
	assert.Equal(t, []int{3}, calls)
}
