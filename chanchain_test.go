package chanchain_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Nomango/chanchain"
)

func TestChain(t *testing.T) {
	ch := make(chan int)
	s := chanchain.NewSource(ch)

	vInt := s.Listen(context.Background())
	vInt.OnChange(func(i interface{}) {
		fmt.Println(i)
	})
	vInt32 := chanchain.Convert(vInt, func(v interface{}) interface{} {
		return int32(v.(int))
	})
	vStr := chanchain.Convert(vInt, func(v interface{}) interface{} {
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
