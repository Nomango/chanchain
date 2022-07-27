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
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})
	vInt32 := react.Convert(vInt, func(v interface{}) interface{} {
		return int32(v.(int))
	})
	vStr := react.Convert(vInt, func(v interface{}) interface{} {
		return fmt.Sprint(v)
	})

	ch <- 1
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, 1, vInt.Load())
	AssertEqual(t, int32(1), vInt32.Load())
	AssertEqual(t, "1", vStr.Load())
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
