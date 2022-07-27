package react

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Source ...
type Source[T any] interface {
	// OnChange registers a handler that handles value changes
	OnChange(f func(T))

	// Change updates the Source
	Change(v T)
}

// NewSource ...
func NewSource[T any](ch <-chan T) Source[T] {
	return &channelSource[T]{ch: ch}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source[time.Time] {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

// Value ...
type Value[T any] interface {
	// Load ...
	Load() T

	// Store ...
	Store(v T)

	// OnChange registers a handler that handles value changes
	OnChange(f func(T))

	// Subscribe receives values from Source and store it into Value
	Subscribe(s Source[T])
}

// NewValue ...
func NewValue[T any]() Value[T] {
	return &value[T]{}
}

// NewValueFrom ...
func NewValueFrom[T any](vv T) Value[T] {
	var v value[T]
	v.v.Store(vv)
	return &v
}

// SubscribeWithTransform receives values from Source and store it into Value with a Transform
func SubscribeWithTransform[T, U any](s Source[T], v Value[U], t Transform[T, U]) {
	s.OnChange(func(vv T) {
		v.Store(t(vv))
	})
}

// Transform converts T to U
type Transform[T, U any] func(T) U

// Bind binds two value with a transform
func Bind[T, U any](from Value[T], to Value[U], t Transform[T, U]) {
	from.OnChange(func(vv T) {
		to.Store(t(vv))
	})
}

// Convert converts Value type
func Convert[T, U any](v Value[T], t Transform[T, U]) Value[U] {
	var newv value[U]
	Bind[T, U](v, &newv, t)
	return &newv
}

type source[T any] struct {
	mu   sync.Mutex
	subs []func(T)
}

func (s *source[T]) OnChange(f func(T)) {
	s.mu.Lock()
	s.subs = append(s.subs, f)
	s.mu.Unlock()
}

func (s *source[T]) Change(vv T) {
	s.mu.Lock()
	subs := s.subs
	s.mu.Unlock()
	if len(subs) > 0 {
		for _, f := range subs {
			go f(vv)
		}
	}
}

type channelSource[T any] struct {
	source[T]
	once sync.Once
	ch   <-chan T
}

func (s *channelSource[T]) OnChange(f func(T)) {
	s.start()
	s.source.OnChange(f)
}

func (s *channelSource[T]) start() {
	s.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetFinalizer(s, func(s *channelSource[T]) {
			cancel()
		})
		go func() {
			for {
				select {
				case vv, ok := <-s.ch:
					if !ok {
						return
					}
					s.Change(vv)
				case <-ctx.Done():
					return
				}
			}
		}()
	})
}

type value[T any] struct {
	source[T]
	v atomic.Value
}

func (v *value[T]) Load() T {
	r, _ := v.v.Load().(T)
	return r
}

func (v *value[T]) Store(vv T) {
	v.v.Store(vv)
	v.Change(vv)
}

func (v *value[T]) Subscribe(s Source[T]) {
	s.OnChange(func(vv T) {
		v.Store(vv)
	})
}
