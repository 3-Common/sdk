//go:build go1.23

package pagination

import "iter"

// All returns a range-over-func iterator that yields each value plus any
// terminal error. Available only on Go 1.23+; the callback API ([Iter.Next] /
// [Iter.Current] / [Iter.Err]) is supported on every release.
//
//	for ev, err := range api.Events.ListAutoPaginate(ctx, nil).All() {
//		if err != nil {
//			return err
//		}
//		_ = ev
//	}
//
// The error slot is non-nil only at the terminal yield, when the underlying
// pagination call fails. Successful yields always pair a value with nil.
func (it *Iter[T]) All() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for it.Next() {
			if !yield(it.current, nil) {
				return
			}
		}
		if err := it.err; err != nil {
			var zero T
			yield(zero, err)
		}
	}
}
