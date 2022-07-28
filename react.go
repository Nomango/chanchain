package react

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Source ...
type Source interface {
	// Change updates the Source
	Change(v interface{})

	// OnChange registers a handler that handles value changes
	OnChange(f func(interface{})) CancelFunc
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

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
	// Load ...
	Load() interface{}

	// Store ...
	Store(v interface{})

	// OnChange registers a handler that handles value changes
	OnChange(f func(interface{})) CancelFunc

	// Bind binds two value with a transform
	Bind(from Value, t Transform) CancelFunc

	// Subscribe receives values from Source and store it into Value
	Subscribe(s Source) CancelFunc

	// SubscribeWithTransform receives values from Source and store it into Value with a Transform
	SubscribeWithTransform(s Source, t Transform) CancelFunc
}

// NewValue ...
func NewValue() Value {
	return &value{}
}

// NewValueFrom ...
func NewValueFrom(vv interface{}) Value {
	var v value
	v.v.Store(vv)
	return &v
}

// Transform converts interface{} to interface{}
type Transform func(interface{}) interface{}

// Convert converts Value type
func Convert(v Value, t Transform) Value {
	var newv value
	newv.Bind(v, t)
	return &newv
}

type subscription func(interface{})

type source struct {
	mu   sync.Mutex
	subs map[*subscription]struct{}
}

func (s *source) Change(vv interface{}) {
	s.mu.Lock()
	subs := s.subs
	s.mu.Unlock()
	if len(subs) > 0 {
		for sub := range subs {
			go (*sub)(vv)
		}
	}
}

func (s *source) OnChange(f func(interface{})) CancelFunc {
	s.mu.Lock()
	sub := (*subscription)(&f)
	if s.subs == nil {
		s.subs = make(map[*subscription]struct{})
	}
	s.subs[sub] = struct{}{}
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.subs, sub)
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

func (v *value) Bind(from Value, t Transform) CancelFunc {
	return from.OnChange(func(vv interface{}) {
		v.Store(t(vv))
	})
}

func (v *value) Subscribe(s Source) CancelFunc {
	return s.OnChange(v.Store)
}

func (v *value) SubscribeWithTransform(s Source, t Transform) CancelFunc {
	return s.OnChange(func(vv interface{}) {
		v.Store(t(vv))
	})
}
