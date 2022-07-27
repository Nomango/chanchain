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
	s := react.NewSource(ch)

	vInt := react.NewValue(0)
	vInt.Subscribe(s)
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})

	vInt2 := react.NewValue(0)
	vInt2.Subscribe(s)

	var vInt32 react.Value
	react.Bind(vInt, &vInt32, func(v interface{}) interface{} {
		return int32(v.(int) + 1)
	})

	vStr := react.Convert(vInt, func(v interface{}) interface{} {
		return fmt.Sprint(v.(int) + 2)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 1, vInt2.Load())
	AssertEqual(t, int32(2), vInt32.Load())
	AssertEqual(t, "3", vStr.Load())
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
