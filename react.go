package react

import (
	"sync"
	"sync/atomic"
	"time"
)

// Binding ...
type Binding[T any] interface {
	// OnChange registers a handler f that handles value changes.
	// Note that f should not perform time-consuming operations.
	OnChange(f func(T)) CancelFunc

	// Binding is a helper function to return this
	Binding() Binding[T]
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

// Transform converts T to U
type Transform[T, U any] func(T) U

// NewBinding ...
func NewBinding[T, U any](from Binding[T], t Transform[T, U]) Binding[U] {
	return &binding[T, U]{from: from, transform: t}
}

// NewAsyncBinding creates a binding that runs in a separate goroutine
func NewAsyncBinding[T, U any](from Binding[T], t Transform[T, U]) Binding[U] {
	return &asyncBinding[T, U]{from: from, transform: t}
}

// Source ...
type Source[T any] interface {
	Binding[T]

	// Change updates the Source
	Change(v T)
}

// NewSource ...
func NewSource[T any]() Source[T] {
	return &source[T]{}
}

// NewChanSource ...
func NewChanSource[T any](ch <-chan T) Source[T] {
	return &channelSource[T]{ch: ch}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source[time.Time] {
	t := time.NewTicker(interval)
	return NewChanSource(t.C)
}

// Value ...
type Value[T any] interface {
	Binding[T]

	// Load ...
	Load() T

	// Store ...
	Store(v T)

	// Bind creates a binding to this value
	Bind(b Binding[T]) CancelFunc
}

// NewValue ...
func NewValue[T any]() Value[T] {
	return &value[T]{}
}

// NewValueFrom ...
func NewValueFrom[T any](vv T) Value[T] {
	v := NewValue[T]()
	v.Store(vv)
	return v
}

// NewBindingValue ...
func NewBindingValue[T any](b Binding[T]) (Value[T], CancelFunc) {
	newv := NewValue[T]()
	return newv, newv.Bind(b)
}

type binding[T, U any] struct {
	from      Binding[T]
	transform Transform[T, U]
}

func (b *binding[T, U]) OnChange(to func(U)) CancelFunc {
	return b.from.OnChange(func(vv T) {
		to(b.transform(vv))
	})
}

func (b *binding[T, U]) Binding() Binding[U] {
	return b
}

type asyncBinding[T, U any] binding[T, U]

func (ab *asyncBinding[T, U]) OnChange(to func(U)) CancelFunc {
	return ab.from.OnChange(func(vv T) {
		go func() {
			to(ab.transform(vv))
		}()
	})
}

func (ab *asyncBinding[T, U]) Binding() Binding[U] {
	return ab
}

type source[T any] struct {
	mu       sync.Mutex
	bindings map[*func(T)]struct{}
}

func (s *source[T]) Binding() Binding[T] {
	return s
}

func (s *source[T]) Change(vv T) {
	s.mu.Lock()
	bindings := s.bindings
	s.mu.Unlock()
	if len(bindings) > 0 {
		for binding := range bindings {
			(*binding)(vv) // not async
		}
	}
}

func (s *source[T]) OnChange(f func(T)) CancelFunc {
	s.mu.Lock()
	fptr := &f
	if s.bindings == nil {
		s.bindings = make(map[*func(T)]struct{})
	}
	s.bindings[fptr] = struct{}{}
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.bindings, fptr)
		s.mu.Unlock()
	}
}

type channelSource[T any] struct {
	source[T]
	once sync.Once
	ch   <-chan T
}

func (s *channelSource[T]) Binding() Binding[T] {
	return s
}

func (s *channelSource[T]) OnChange(f func(T)) CancelFunc {
	s.start()
	return s.source.OnChange(f)
}

func (s *channelSource[T]) start() {
	s.once.Do(func() {
		go func() {
			for {
				vv, ok := <-s.ch
				if !ok {
					return
				}
				s.Change(vv)
			}
		}()
	})
}

type value[T any] struct {
	source[T]
	v atomic.Value
}

func (v *value[T]) Load() T {
	vv, _ := v.v.Load().(T)
	return vv
}

func (v *value[T]) Store(vv T) {
	v.v.Store(vv)
	v.Change(vv)
}

func (v *value[T]) Bind(b Binding[T]) CancelFunc {
	return b.Binding().OnChange(v.Store)
}

func (v *value[T]) Binding() Binding[T] {
	return v
}
