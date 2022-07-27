package react

import (
	"sync"
	"sync/atomic"
	"time"
)

// Value ...
type Value[T any] struct {
	v        atomic.Value
	mu       sync.Mutex
	onChange []func(T)
}

// NewValue ...
func NewValue[T any](vv T) *Value[T] {
	var v Value[T]
	v.v.Store(vv)
	return &v
}

// Load ...
func (v *Value[T]) Load() T {
	r, _ := v.v.Load().(T)
	return r
}

// Store ...
func (v *Value[T]) Store(vv T) {
	v.change(vv)
}

// OnChange registers a handler that handles value changes
func (v *Value[T]) OnChange(f func(T)) {
	v.mu.Lock()
	v.onChange = append(v.onChange, f)
	v.mu.Unlock()
}

// Subscribe receives values from Source and store it into Value
func (v *Value[T]) Subscribe(s Source[T]) {
	(*source[T])(s).start()
	Bind(s.v, v, func(vv T) T { return vv })
}

func (v *Value[T]) change(vv T) {
	v.v.Store(vv)

	v.mu.Lock()
	onChange := v.onChange
	v.mu.Unlock()
	if len(onChange) > 0 {
		go func() {
			for _, f := range onChange {
				f(vv)
			}
		}()
	}
}

// Transform converts T to U
type Transform[T, U any] func(T) U

// Bind binds two value with a transform
func Bind[T, U any](from *Value[T], to *Value[U], t Transform[T, U]) {
	from.OnChange(func(vv T) {
		to.change(t(vv))
	})
}

// Convert converts Value type
func Convert[T, U any](v *Value[U], t Transform[U, T]) *Value[T] {
	var newv Value[T]
	Bind(v, &newv, t)
	return &newv
}

// Source ...
type Source[T any] *source[T]

// NewSource ...
func NewSource[T any](ch <-chan T) Source[T] {
	return &source[T]{ch: ch, v: &Value[T]{}}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source[time.Time] {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

type source[T any] struct {
	ch   <-chan T
	v    *Value[T]
	once sync.Once
}

func (s *source[T]) start() {
	s.once.Do(func() {
		go func() {
			for {
				vv, ok := <-s.ch
				if !ok {
					return
				}
				s.v.change(vv)
			}
		}()
	})
}
