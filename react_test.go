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

	vInt := react.NewValue()
	vInt.Bind(s)
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})

	vInt2 := react.NewValueFrom(0)
	vInt2.Bind(s, react.WithTransform(func(v interface{}) interface{} {
		return v.(int) + 1
	}))

	vInt32 := react.NewValue()
	vInt32.Bind(vInt, react.WithTransform(func(v interface{}) interface{} {
		return int32(v.(int) + 2)
	}))

	vStr := react.NewBindingValue(vInt, react.WithTransform(func(v interface{}) interface{} {
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

	vInt1 := react.NewValue()
	cancel1 := vInt1.Bind(s)

	vInt2 := react.NewValue()
	cancel2 := vInt2.Bind(s)

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
	v2 := react.NewBindingValue(v1)
	v3 := react.NewBindingValue(v1, react.WithTransform(func(i interface{}) interface{} {
		time.Sleep(time.Millisecond * 50)
		return i
	}), react.WithAsync(true))

	v1.Store(1)
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, nil, v3.Load())

	time.Sleep(time.Millisecond * 50)

	AssertEqual(t, 1, v2.Load())
	AssertEqual(t, 1, v3.Load())
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("\n%v:%v:\n\tvalues are not equal\n\texpected=%v\n\tgot=%v", file, line, expect, actual)
	}
}
