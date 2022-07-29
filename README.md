# go-react

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for earlier version of Golang](https://github.com/Nomango/go-react/tree/legacy).

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewChanSource(ch)

// Create a value and bind with the source
vInt := react.NewValue[int]()
cancel := vInt.Bind(s)

// A binding can be canceled
defer cancel()

// Set action on change
vInt.OnChange(func(i int) {
    fmt.Println(i)
})

// A source can be bound more than one time
// So the following code is valid
vInt2 := react.NewValueFrom(0)
vInt2.Bind(react.NewBinding(s.Binding(), func(v int) int {
    return v + 1 // Processing raw value
}))

// Bind another value
vInt32 := react.NewValue[int32]()
vInt32.Bind(react.NewBinding(vInt.Binding(), func(v int) int32 {
    return int32(v + 1)
}))

// Convert a int value to a string value
asyncBinding := react.NewAsyncBinding(vInt.Binding(), func(v int) string {
    return fmt.Sprint(v + 3) // Processing in a separate goroutine
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
