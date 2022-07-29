# go-react

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for Golang1.18+](https://github.com/Nomango/go-react).

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewChanSource(ch)

// Create a value and bind the source
vInt := react.NewValue()
vInt.Bind(s)

// A source can be bound more than one time
// So the following code is valid
vInt2 := react.NewValueFrom(0)
cancel := vInt2.Bind(s)

// A binding can be canceled
cancel()

// Set action on change
vInt.OnChange(func(i interface{}) {
    fmt.Println(i)
})

// Bind another value with a transform
vInt32 := react.NewValue()
vInt32.Bind(vInt, react.WithTransform(func(i interface{}) interface{} {
    return int32(i.(int) + 1)
}))

// Convert a int value to a string value
vStr := react.NewBindingValue(vInt, react.WithTransform(func(i interface{}) interface{} {
    return fmt.Sprint(i.(int) + 2)
}))

// Send a value to source
ch <- 1

// Wait for the update to complete
time.Sleep(time.Millisecond * 10)

fmt.Println(vInt32.Load())
fmt.Println(vStr.Load())

// Output:
// 1
// 2
// 3
```
