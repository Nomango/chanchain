package react

import (
	"sync"
	"sync/atomic"
	"time"
)

// Source ...
type Source[T any] interface {
	// Change updates the Source
	Change(v T)

	// OnChange registers a handler that handles value changes
	OnChange(f func(T)) CancelFunc
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

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
	// Load ...
	Load() T

	// Store ...
	Store(v T)

	// OnChange registers a handler that handles value changes
	OnChange(f func(T)) CancelFunc

	// Subscribe receives values from Source and store it into Value
	Subscribe(s Source[T]) CancelFunc
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
func SubscribeWithTransform[T, U any](s Source[T], v Value[U], t Transform[T, U]) CancelFunc {
	return s.OnChange(func(vv T) {
		v.Store(t(vv))
	})
}

// Transform converts T to U
type Transform[T, U any] func(T) U

// Bind binds two value with a transform
func Bind[T, U any](from Value[T], to Value[U], t Transform[T, U]) CancelFunc {
	return from.OnChange(func(vv T) {
		to.Store(t(vv))
	})
}

// Convert converts Value type
func Convert[T, U any](v Value[T], t Transform[T, U]) Value[U] {
	var newv value[U]
	Bind[T, U](v, &newv, t)
	return &newv
}

type subscription[T any] func(T)

type source[T any] struct {
	mu   sync.Mutex
	subs map[*subscription[T]]struct{}
}

func (s *source[T]) Change(vv T) {
	s.mu.Lock()
	subs := s.subs
	s.mu.Unlock()
	if len(subs) > 0 {
		for sub := range subs {
			go (*sub)(vv)
		}
	}
}

func (s *source[T]) OnChange(f func(T)) CancelFunc {
	s.mu.Lock()
	sub := (*subscription[T])(&f)
	if s.subs == nil {
		s.subs = make(map[*subscription[T]]struct{})
	}
	s.subs[sub] = struct{}{}
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.subs, sub)
		s.mu.Unlock()
	}
}

type channelSource[T any] struct {
	source[T]
	once sync.Once
	ch   <-chan T
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
	r, _ := v.v.Load().(T)
	return r
}

func (v *value[T]) Store(vv T) {
	v.v.Store(vv)
	v.Change(vv)
}

func (v *value[T]) Subscribe(s Source[T]) CancelFunc {
	return s.OnChange(v.Store)
}
