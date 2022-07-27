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
	vInt2.Subscribe(s)

	vInt32 := react.NewValue[int32]()
	react.Bind(vInt, vInt32, func(v int) int32 {
		return int32(v + 1)
	})

	vStr := react.Convert(vInt, func(v int) string {
		return fmt.Sprint(v + 2)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 1, vInt2.Load())
	AssertEqual(t, 2, vInt32.Load())
	AssertEqual(t, "3", vStr.Load())
}

func AssertEqual[T any](t *testing.T, expect, actual T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
