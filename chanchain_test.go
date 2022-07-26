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
	c := chanchain.NewChain(func(v interface{}) interface{} {
		return fmt.Sprint(v) + "1"
	}, func(v interface{}) interface{} {
		return fmt.Sprint(v) + "2"
	})

	c.Append(func(v interface{}) interface{} {
		return fmt.Sprint(v) + "3"
	}, func(v interface{}) interface{} {
		return fmt.Sprint(v) + "4"
	})

	ch := make(chan int)
	s := chanchain.NewSource(ch)
	c.Start(context.Background(), s)

	v := chanchain.NewValue(c)

	ch <- 0
	time.Sleep(time.Millisecond * 10)

	AssertEqual(t, "01234", v.Load())
}

func AssertEqual(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		t.Fatalf("values are not equal\nexpected=%v\ngot=%v", expect, actual)
	}
}
