// Package pagination contains the generic auto-paginating iterator returned
// by every list endpoint's auto-paginate variant.
package pagination

// Iter generic auto-paginating iterator returned by every list endpoint's
// auto-paginate variant. It walks pages lazily — one HTTP call per page, only
// when the consumer drains the previous page's buffer.
//
// Idiomatic usage:
//
//	iter := api.Events.ListAutoPaginate(ctx, nil)
//	for iter.Next() {
//		ev := iter.Current()
//		_ = ev
//	}
//	if err := iter.Err(); err != nil {
//		// non-nil if any page failed
//	}
//
// On Go 1.23+ a range-over-func variant is also supported — see [Iter.All].
type Iter[T any] struct {
	fetchPage func(page int) ([]T, bool, error)
	page      int
	buffer    []T
	index     int
	hasMore   bool
	current   T
	err       error
}

// NewIter constructs an [*Iter] backed by fetchPage. fetchPage receives a
// 0-indexed page number and returns (data, hasMore, err). Not part of the
// public API; resource packages call this from their list-auto-paginate
// methods.
//
//nolint:revive // referenced from resource packages
func NewIter[T any](startPage int, fetchPage func(page int) ([]T, bool, error)) *Iter[T] {
	return &Iter[T]{
		fetchPage: fetchPage,
		page:      startPage,
		hasMore:   true,
	}
}

// Next advances the iterator. Returns true on success, false when the stream
// is exhausted or an error occurred. After Next returns false, callers should
// check [Iter.Err] to distinguish end-of-stream from failure.
func (it *Iter[T]) Next() bool {
	if it.err != nil {
		return false
	}

	if it.index < len(it.buffer) {
		it.current = it.buffer[it.index]
		it.index++
		return true
	}

	if !it.hasMore {
		return false
	}

	data, hasMore, err := it.fetchPage(it.page)
	if err != nil {
		it.err = err
		return false
	}
	it.buffer = data
	it.index = 0
	it.hasMore = hasMore
	it.page++

	if len(it.buffer) == 0 {
		return false
	}

	it.current = it.buffer[it.index]
	it.index++
	return true
}

// Current returns the most recent value yielded by [Iter.Next]. Calling
// Current before the first successful Next, or after Next has returned false,
// returns the zero value of T.
func (it *Iter[T]) Current() T { return it.current }

// Err returns the first error encountered during iteration, or nil if
// iteration completed cleanly.
func (it *Iter[T]) Err() error { return it.err }
