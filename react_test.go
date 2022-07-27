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

	vInt := react.NewValue[int]()
	vInt.Subscribe(s)
	vInt.OnChange(func(i int) {
		fmt.Println(i)
	})

	vInt2 := react.NewValueFrom(0)
	react.SubscribeWithTransform(s, vInt2, func(v int) int {
		return v + 1
	})

	vInt32 := react.NewValue[int32]()
	react.Bind(vInt, vInt32, func(v int) int32 {
		return int32(v + 2)
	})

	vStr := react.Convert(vInt, func(v int) string {
		return fmt.Sprint(v + 3)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 2, vInt2.Load())
	AssertEqual(t, 3, vInt32.Load())
	AssertEqual(t, "4", vStr.Load())
}

func AssertEqual[T any](t *testing.T, expect, actual T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
