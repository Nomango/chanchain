package react_test

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Nomango/go-react"
)

func TestReact(t *testing.T) {
	ch := make(chan int)
	s := react.NewChanSource(ch)

	vInt := react.NewValue()
	vInt.Bind(s)
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})

	vInt2 := react.NewValueFrom(0)
	vInt2.Bind(react.NewBinding(s, func(v interface{}) interface{} {
		return v.(int) + 1
	}))

	vInt32, _ := react.NewBindingValue(react.NewBinding(vInt, func(v interface{}) interface{} {
		return int32(v.(int) + 2)
	}))

	vStr, _ := react.NewBindingValue(react.NewAsyncBinding(vInt, func(v interface{}) interface{} {
		return fmt.Sprint(v.(int) + 3)
	}))

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 2, vInt2.Load())
	AssertEqual(t, int32(3), vInt32.Load())
	AssertEqual(t, "4", vStr.Load())
}

func TestCancel(t *testing.T) {
	s := react.NewSource()
	vInt1, cancel1 := react.NewBindingValue(s)
	vInt2, cancel2 := react.NewBindingValue(s)

	s.Change(1)
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt1.Load())
	AssertEqual(t, 1, vInt2.Load())

	cancel1()
	s.Change(2)
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt1.Load())
	AssertEqual(t, 2, vInt2.Load())

	cancel2()
	s.Change(3)
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt1.Load())
	AssertEqual(t, 2, vInt2.Load())
}

func TestBlock(t *testing.T) {
	v1 := react.NewValueFrom(0)
	v2, _ := react.NewBindingValue(v1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	v3, _ := react.NewBindingValue(react.NewAsyncBinding(v1, func(i interface{}) interface{} {
		defer wg.Done()
		time.Sleep(time.Millisecond * 100)
		return i
	}))

	v1.Store(1)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, nil, v3.Load())

	wg.Wait()
	time.Sleep(time.Millisecond * 100)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, 1, v3.Load())
}

func TestTickSource(t *testing.T) {
	interval := time.Millisecond * 200
	s := react.NewTickSource(interval)
	v, _ := react.NewBindingValue(s)

	time.Sleep(interval * 2)

	AssertNotEqual(t, nil, v.Load())
}

func TestCloseChanSource(t *testing.T) {
	ch := make(chan int)
	s := react.NewChanSource(ch)
	s.OnChange(func(i interface{}) {})

	close(ch)
	time.Sleep(time.Millisecond * 50)
}

func TestError(t *testing.T) {
	AssertPanic(t, func() {
		s := react.NewChanSource(1)
		s.OnChange(func(i interface{}) {})
	})
	AssertPanic(t, func() {
		ch := make(chan int)
		s := react.NewChanSource((chan<- int)(ch))
		s.OnChange(func(i interface{}) {})
	})
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%v:%v:\n\tValues are not equal\n\texpected=%#v\n\tgot=%#v", file, line, expect, actual)
	}
}

func AssertNotEqual(t *testing.T, expect, actual interface{}) {
	if reflect.DeepEqual(expect, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%v:%v:\n\tShould not be: %#v", file, line, actual)
	}
}

func AssertPanic(t *testing.T, f func()) {
	defer func() {
		if e := recover(); e == nil {
			_, file, line, _ := runtime.Caller(2)
			t.Fatalf("\n%v:%v:\n\tShould panic", file, line)
		}
	}()
	f()
}
