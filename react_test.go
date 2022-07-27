package react_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Nomango/react"
)

func TestReact(t *testing.T) {
	ch := make(chan int)
	s := react.NewChanSource(ch)

	vInt := react.NewValue()
	vInt.Subscribe(s)
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})

	vInt2 := react.NewValueFrom(0)
	vInt2.SubscribeWithTransform(s, func(v interface{}) interface{} {
		return v.(int) + 1
	})

	vInt32 := react.NewValue()
	vInt32.Bind(vInt, func(v interface{}) interface{} {
		return int32(v.(int) + 2)
	})

	vStr := react.Convert(vInt, func(v interface{}) interface{} {
		return fmt.Sprint(v.(int) + 3)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 2, vInt2.Load())
	AssertEqual(t, int32(3), vInt32.Load())
	AssertEqual(t, "4", vStr.Load())
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
