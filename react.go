package react

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Source ...
type Source interface {
	// OnChange registers a handler that handles value changes
	OnChange(f func(interface{}))

	// Change updates the Source
	Change(v interface{})
}

// NewSource ...
func NewSource(ch interface{}) Source {
	return &channelSource{ch: wrapChannel(ch)}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

// Value ...
type Value interface {
	// Load ...
	Load() interface{}

	// Store ...
	Store(v interface{})

	// OnChange registers a handler that handles value changes
	OnChange(f func(interface{}))

	// Bind binds two value with a transform
	Bind(from Value, t Transform)

	// Subscribe receives values from Source and store it into Value
	Subscribe(s Source)

	// SubscribeWithTransform receives values from Source and store it into Value with a Transform
	SubscribeWithTransform(s Source, t Transform)
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

type source struct {
	mu   sync.Mutex
	subs []func(interface{})
}

func (s *source) OnChange(f func(interface{})) {
	s.mu.Lock()
	s.subs = append(s.subs, f)
	s.mu.Unlock()
}

func (s *source) Change(vv interface{}) {
	s.mu.Lock()
	subs := s.subs
	s.mu.Unlock()
	if len(subs) > 0 {
		for _, f := range subs {
			go f(vv)
		}
	}
}

type channelSource struct {
	source
	once sync.Once
	ch   <-chan interface{}
}

func (s *channelSource) OnChange(f func(interface{})) {
	s.start()
	s.source.OnChange(f)
}

func (s *channelSource) start() {
	s.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetFinalizer(s, func(s *channelSource) {
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

func (v *value) Bind(from Value, t Transform) {
	from.OnChange(func(vv interface{}) {
		v.Store(t(vv))
	})
}

func (v *value) Subscribe(s Source) {
	s.OnChange(func(vv interface{}) {
		v.Store(vv)
	})
}

func (v *value) SubscribeWithTransform(s Source, t Transform) {
	s.OnChange(func(vv interface{}) {
		v.Store(t(vv))
	})
}

func wrapChannel(ch interface{}) <-chan interface{} {
	t := reflect.TypeOf(ch)
	if t.Kind() != reflect.Chan || t.ChanDir()&reflect.RecvDir == 0 {
		panic("channels: input to Wrap must be readable channel")
	}
	realChan := make(chan interface{})

	go func() {
		v := reflect.ValueOf(ch)
		for {
			x, ok := v.Recv()
			if !ok {
				close(realChan)
				return
			}
			realChan <- x.Interface()
		}
	}()
	return realChan
}
