package react

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Binding ...
type Binding interface {
	// OnChange registers a handler f that handles value changes.
	// Note that f should not perform time-consuming operations.
	OnChange(f func(interface{})) CancelFunc
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

// Transform converts interface{} to interface{}
type Transform func(interface{}) interface{}

// NewBinding ...
func NewBinding(from Binding, t Transform) Binding {
	return &binding{from: from, transform: t}
}

// NewAsyncBinding creates a binding that runs in a separate goroutine
func NewAsyncBinding(from Binding, t Transform) Binding {
	return &asyncBinding{from: from, transform: t}
}

// Source ...
type Source interface {
	Binding

	// Change updates the Source
	Change(v interface{})
}

// NewSource ...
func NewSource() Source {
	return &source{}
}

// NewChanSource ...
func NewChanSource(ch interface{}) Source {
	return &channelSource{ch: ch}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source {
	t := time.NewTicker(interval)
	return NewChanSource(t.C)
}

// Value ...
type Value interface {
	Binding

	// Load ...
	Load() interface{}

	// Store ...
	Store(v interface{})

	// Bind creates a binding to this value
	Bind(b Binding) CancelFunc
}

// NewValue ...
func NewValue() Value {
	return &value{}
}

// NewValueFrom ...
func NewValueFrom(vv interface{}) Value {
	v := NewValue()
	v.Store(vv)
	return v
}

// NewBindingValue ...
func NewBindingValue(b Binding) (Value, CancelFunc) {
	newv := NewValue()
	return newv, newv.Bind(b)
}

type binding struct {
	from      Binding
	transform Transform
}

func (b *binding) OnChange(to func(interface{})) CancelFunc {
	return b.from.OnChange(func(vv interface{}) {
		to(b.transform(vv))
	})
}

type asyncBinding binding

func (ab *asyncBinding) OnChange(to func(interface{})) CancelFunc {
	return ab.from.OnChange(func(vv interface{}) {
		go func() {
			to(ab.transform(vv))
		}()
	})
}

type source struct {
	mu       sync.Mutex
	bindings map[*func(interface{})]struct{}
}

func (s *source) Change(vv interface{}) {
	s.mu.Lock()
	bindings := s.bindings
	s.mu.Unlock()
	if len(bindings) > 0 {
		for binding := range bindings {
			(*binding)(vv) // not async
		}
	}
}

func (s *source) OnChange(f func(interface{})) CancelFunc {
	s.mu.Lock()
	fptr := &f
	if s.bindings == nil {
		s.bindings = make(map[*func(interface{})]struct{})
	}
	s.bindings[fptr] = struct{}{}
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.bindings, fptr)
		s.mu.Unlock()
	}
}

type channelSource struct {
	source
	once sync.Once
	ch   interface{}
}

func (s *channelSource) OnChange(f func(interface{})) CancelFunc {
	s.start()
	return s.source.OnChange(f)
}

func (s *channelSource) start() {
	s.once.Do(func() {
		t := reflect.TypeOf(s.ch)
		if t.Kind() != reflect.Chan || t.ChanDir()&reflect.RecvDir == 0 {
			panic("go-react: input to ChanSource must be readable channel")
		}
		v := reflect.ValueOf(s.ch)
		go func() {
			for {
				vv, ok := v.Recv()
				if !ok {
					return
				}
				s.Change(vv.Interface())
			}
		}()
	})
}

type value struct {
	source
	v atomic.Value
}

func (v *value) Load() interface{} {
	return v.v.Load()
}

func (v *value) Store(vv interface{}) {
	v.v.Store(vv)
	v.Change(vv)
}

func (v *value) Bind(b Binding) CancelFunc {
	return b.OnChange(v.Store)
}
