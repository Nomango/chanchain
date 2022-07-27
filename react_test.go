package react_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Nomango/react"
)

func TestReact(t *testing.T) {
	ch := make(chan int)
	s := react.NewSource(ch)

	vInt := s.Subscribe(context.Background())
	vInt32 := react.Convert(vInt, func(v int) int32 {
		return int32(v)
	})
	vStr := react.Convert(vInt, func(v int) string {
		return fmt.Sprint(v)
	})
	vStr.OnChange(func(s string) {
		fmt.Println(s)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, 1, vInt32.Load())
	AssertEqual(t, "1", vStr.Load())
}

func AssertEqual[T any](t *testing.T, expect, actual T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
