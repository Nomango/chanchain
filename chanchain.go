package chanchain

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Source[T any] <-chan T

func NewSource[T any](ch <-chan T) Source[T] {
	return Source[T](ch)
}

func NewTickSource(interval time.Duration) Source[time.Time] {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

// Listen receives values from Source and store it into Value
func (s Source[T]) Listen(ctx context.Context) *Value[T] {
	var v Value[T]
	go func() {
		for {
			select {
			case vv, ok := <-s:
				if !ok {
					return
				}
				v.change(vv)
			case <-ctx.Done():
				return
			}
		}
	}()
	return &v
}

type Value[T any] struct {
	v        atomic.Value
	mu       sync.Mutex
	onChange []func(T)
}

func (v *Value[T]) Load() T {
	r, _ := v.v.Load().(T)
	return r
}

func (v *Value[T]) OnChange(f func(T)) {
	v.mu.Lock()
	v.onChange = append(v.onChange, f)
	v.mu.Unlock()
}

func (v *Value[T]) change(vv T) {
	v.v.Store(vv)

	v.mu.Lock()
	onChange := v.onChange
	v.mu.Unlock()
	if len(onChange) > 0 {
		for _, f := range onChange {
			f := f
			go f(vv)
		}
	}
}

type Transform[T, U any] func(T) U

func Convert[T, U any](v *Value[U], t Transform[U, T]) *Value[T] {
	var newv Value[T]
	v.OnChange(func(u U) {
		newv.change(t(u))
	})
	return &newv
}
