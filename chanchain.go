package chanchain

import (
	"context"
	"reflect"
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

type Transform func(interface{}) interface{}

type Chain struct {
	v atomic.Value
}

func NewChain(ts ...Transform) *Chain {
	var c Chain
	c.Append(ts...)
	return &c
}

func (c *Chain) Start(ctx context.Context, s Source) {
	go func() {
		for {
			select {
			case v, ok := <-s:
				if !ok {
					return
				}
				chain := c.v.Load().(*chain)
				_ = (*chain)(func(v interface{}) interface{} { return v })(v)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Chain) Append(ts ...Transform) {
	if len(ts) == 0 {
		return
	}
	var t Transform
	if len(ts) == 1 {
		t = ts[0]
	} else {
		t = func(v interface{}) interface{} {
			for _, t := range ts {
				v = t(v)
			}
			return v
		}
	}
	var newc chain
	if prev, _ := c.v.Load().(*chain); prev == nil {
		newc = func(next Transform) Transform {
			return func(v interface{}) interface{} {
				return next(t(v))
			}
		}
	} else {
		newc = func(next Transform) Transform {
			return func(v interface{}) interface{} {
				return next((*prev)(t)(v))
			}
		}
	}
	c.v.Store(&newc)
}

type Value struct {
	v atomic.Value
}

func NewValue(c *Chain) *Value {
	var value Value
	c.Append(func(v interface{}) interface{} {
		value.v.Store(v)
		return v
	})
	return &value
}

func (v *Value) Load() interface{} {
	return v.v.Load()
}

type chain func(Transform) Transform

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
