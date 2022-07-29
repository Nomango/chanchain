package react_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/Nomango/react"
)

func TestReact(t *testing.T) {
	ch := make(chan int)
	s := react.NewChanSource(ch)

	vInt := react.NewValue[int]()
	vInt.Bind(s)
	vInt.OnChange(func(i int) {
		fmt.Println(i)
	})

	vInt2 := react.NewValueFrom(0)
	vInt2.Bind(react.NewBinding(s.Binding(), func(v int) int {
		return v + 1
	}))

	vInt32, _ := react.NewBindingValue(react.NewBinding(vInt.Binding(), func(v int) int32 {
		return int32(v + 2)
	}))

	vStr, _ := react.NewBindingValue(react.NewAsyncBinding(vInt.Binding(), func(v int) string {
		return fmt.Sprint(v + 3)
	}))

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 2, vInt2.Load())
	AssertEqual(t, 3, vInt32.Load())
	AssertEqual(t, "4", vStr.Load())
}

func TestCancel(t *testing.T) {
	s := react.NewSource[int]()

	vInt1, cancel1 := react.NewBindingValue(s.Binding())

	vInt2, cancel2 := react.NewBindingValue(s.Binding())

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
	v2, _ := react.NewBindingValue(v1.Binding())
	v3, _ := react.NewBindingValue(react.NewAsyncBinding(v1.Binding(), func(i int) int {
		time.Sleep(time.Millisecond * 50)
		return i
	}))

	v1.Store(1)
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, 0, v3.Load())

	time.Sleep(time.Millisecond * 50)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, 1, v3.Load())
}

func AssertEqual[T any](t *testing.T, expect, actual T) {
	if !reflect.DeepEqual(expect, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%v:%v:\n\tvalues are not equal\n\texpected=%v\n\tgot=%v", file, line, expect, actual)
	}
}
