package react

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Changeable ...
type Changeable interface {
	// OnChange registers a handler that handles value changes. Note that b should not
	// preform time-consuming operations
	OnChange(b Binding) CancelFunc
}

// Binding ...
type Binding func(interface{})

// NewBinding ...
func NewBinding(f func(interface{}), opts ...BindOption) Binding {
	var bm bindingMaker
	for _, opt := range opts {
		opt(&bm)
	}
	return bm.make(f)
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

// Source ...
type Source interface {
	Changeable

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
	Changeable

	// Load ...
	Load() interface{}

	// Store ...
	Store(v interface{})

	// Bind creates a binding between c and this value
	Bind(c Changeable, opts ...BindOption) CancelFunc
}

// NewValue ...
func NewValue() Value {
	return &value{}
}

// NewValueFrom ...
func NewValueFrom(vv interface{}) Value {
	v := NewValue()
	if vv != nil {
		v.Store(vv)
	}
	return v
}

// NewBindingValue ...
func NewBindingValue(from Value, opts ...BindOption) Value {
	newv := NewValue()
	newv.Bind(from, opts...)
	return newv
}

// BindOption ...
type BindOption func(*bindingMaker)

// Transform converts interface{} to interface{}
type Transform func(interface{}) interface{}

// WithTransform returns an option that indicates the binding should take a Transform
func WithTransform(t Transform) BindOption {
	return func(o *bindingMaker) {
		o.transform = t
	}
}

// WithAsync returns an option that indicates the binding should run in a separate goroutine
func WithAsync(async bool) BindOption {
	return func(o *bindingMaker) {
		o.async = async
	}
}

type source struct {
	mu       sync.Mutex
	bindings map[*Binding]struct{}
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

func (s *source) OnChange(b Binding) CancelFunc {
	s.mu.Lock()
	bptr := &b
	if s.bindings == nil {
		s.bindings = make(map[*Binding]struct{})
	}
	s.bindings[bptr] = struct{}{}
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.bindings, bptr)
		s.mu.Unlock()
	}
}

type channelSource struct {
	source
	once sync.Once
	ch   interface{}
}

func (s *channelSource) OnChange(b Binding) CancelFunc {
	s.start()
	return s.source.OnChange(b)
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

func (v *value) Bind(c Changeable, opts ...BindOption) CancelFunc {
	return c.OnChange(NewBinding(v.Store, opts...))
}

type bindingMaker struct {
	transform Transform
	async     bool
}

func (bm *bindingMaker) make(f func(vv interface{})) Binding {
	if bm.transform != nil {
		prev := f
		f = func(vv interface{}) {
			prev(bm.transform(vv))
		}
	}
	if bm.async {
		prev := f
		f = func(vv interface{}) {
			go prev(vv)
		}
	}
	return f
}
