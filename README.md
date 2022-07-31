# go-react

[![Go Reference](https://pkg.go.dev/badge/github.com/Nomango/go-react.svg)](https://pkg.go.dev/github.com/Nomango/go-react)
[![Github status](https://github.com/Nomango/go-react/actions/workflows/UnitTest.yml/badge.svg?branch=main)](https://github.com/Nomango/go-react/actions)
[![GitHub release](https://img.shields.io/github/release/nomango/go-react)](https://github.com/Nomango/go-react/releases/latest)
[![Codacy Grade](https://app.codacy.com/project/badge/Grade/eccf7bb1fe7144458a76e3905dd4d054)](https://www.codacy.com/gh/Nomango/go-react/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=Nomango/go-react&amp;utm_campaign=Badge_Grade)
[![Codacy Coverage](https://app.codacy.com/project/badge/Coverage/9dd32aab87364903a57f32543f3bf738)](https://www.codacy.com/gh/Nomango/go-react/dashboard?utm_source=github.com&utm_medium=referral&utm_content=Nomango/go-react&utm_campaign=Badge_Coverage)
[![License](https://img.shields.io/github/license/nomango/go-react)](https://github.com/Nomango/go-react/blob/main/LICENSE)

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
