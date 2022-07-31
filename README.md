# go-react

[![Go Reference](https://pkg.go.dev/badge/github.com/Nomango/go-react.svg)](https://pkg.go.dev/github.com/Nomango/go-react/v1)
[![Github status](https://github.com/Nomango/go-react/actions/workflows/UnitTest.yml/badge.svg?branch=1.x)](https://github.com/Nomango/go-react/actions)
[![GitHub release](https://img.shields.io/github/release/nomango/go-react)](https://github.com/Nomango/go-react/releases/latest)
[![codecov](https://codecov.io/gh/Nomango/go-react/branch/main/graph/badge.svg?token=YEGAFMRM28)](https://codecov.io/gh/Nomango/go-react)
[![License](https://img.shields.io/github/license/nomango/go-react)](https://github.com/Nomango/go-react/blob/1.x/LICENSE)

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for Golang1.18+](https://github.com/Nomango/go-react).

## Install

```bash
go get github.com/Nomango/go-react@v1
```

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewChanSource(ch)

// Create a value and bind with the source
vInt := react.NewValue()
cancel := vInt.Bind(s)

// A binding can be canceled
defer cancel()

// Set action on change
vInt.OnChange(func(i interface{}) {
    fmt.Println(i)
})

// A source can be bound more than one time
// So the following code is valid
vInt2 := react.NewValueFrom(0)
vInt2.Bind(react.NewBinding(s, func(v interface{}) interface{} {
    return v.(int) + 1 // Processing raw value
}))

// Bind another value
vInt32 := react.NewValue()
vInt32.Bind(react.NewBinding(vInt, func(v interface{}) interface{} {
    return int32(v.(int) + 1)
}))

// Convert a int value to a string value
asyncBinding := react.NewAsyncBinding(vInt, func(v interface{}) interface{} {
    return fmt.Sprint(v.(int) + 3) // Processing in a separate goroutine
})
vStr, _ := react.NewBindingValue(asyncBinding)

// Send a value to Source
ch <- 1

// Wait for the update to complete
time.Sleep(time.Millisecond * 10)

fmt.Println(vInt2.Load())
fmt.Println(vInt32.Load())
fmt.Println(vStr.Load())

// Output:
// 1
// 2
// 3
// 4
```
