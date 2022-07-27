package react

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Value ...
type Value struct {
	v        atomic.Value
	mu       sync.Mutex
	onChange []func(interface{})
}

// NewValue ...
func NewValue(vv interface{}) *Value {
	var v Value
	v.v.Store(vv)
	return &v
}

// Load ...
func (v *Value) Load() interface{} {
	return v.v.Load()
}

// Store ...
func (v *Value) Store(vv interface{}) {
	v.change(vv)
}

// OnChange registers a handler that handles value changes
func (v *Value) OnChange(f func(interface{})) {
	v.mu.Lock()
	v.onChange = append(v.onChange, f)
	v.mu.Unlock()
}

// Bind binds to another value with the transform
func (v *Value) Bind(other *Value, t Transform) {
	other.OnChange(func(u interface{}) {
		v.change(t(u))
	})
}

// Subscribe receives values from Source and store it into Value
func (v *Value) Subscribe(s Source) {
	(*source)(s).start()
	Bind(s.v, v, func(vv interface{}) interface{} { return vv })
}

func (v *Value) change(vv interface{}) {
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
type Transform func(interface{}) interface{}

// Bind binds two value with a transform
func Bind(from *Value, to *Value, t Transform) {
	from.OnChange(func(vv interface{}) {
		to.change(t(vv))
	})
}

// Convert converts Value type
func Convert(v *Value, t Transform) *Value {
	var newv Value
	Bind(v, &newv, t)
	return &newv
}

// Source ...
type Source *source

// NewSource ...
func NewSource(ch interface{}) Source {
	return &source{ch: wrapChannel(ch), v: &Value{}}
}

// NewTickSource creates a Source that will send the current time after each tick
func NewTickSource(interval time.Duration) Source {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

type source struct {
	ch   <-chan interface{}
	v    *Value
	once sync.Once
}

func (s *source) start() {
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
