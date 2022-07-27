package chanchain

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type Source <-chan interface{}

func NewSource(ch interface{}) Source {
	return Source(wrapChannel(ch))
}

func NewTickSource(interval time.Duration) Source {
	t := time.NewTicker(interval)
	return NewSource(t.C)
}

// Listen receives values from Source and store it into Value
func (s Source) Listen(ctx context.Context) *Value {
	var v Value
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

type Value struct {
	v        atomic.Value
	mu       sync.Mutex
	onChange []func(interface{})
}

func (v *Value) Load() interface{} {
	return v.v.Load()
}

func (v *Value) OnChange(f func(interface{})) {
	v.mu.Lock()
	v.onChange = append(v.onChange, f)
	v.mu.Unlock()
}

func (v *Value) change(vv interface{}) {
	v.v.Store(vv)

	v.mu.Lock()
	fs := v.onChange
	v.mu.Unlock()
	if len(fs) > 0 {
		for _, f := range fs {
			f := f
			go f(vv)
		}
	}
}

type Transform func(interface{}) interface{}

func Convert(v *Value, t Transform) *Value {
	var newv Value
	v.OnChange(func(u interface{}) {
		newv.change(t(u))
	})
	return &newv
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
