# go-react

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for Golang1.18+](https://github.com/Nomango/go-react).

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewChanSource(ch)

// Create a value and subscribe the source
vInt := react.NewValue()
vInt.Subscribe(s)

// A source can be subscribed more than one time
// So the following code is valid
vInt2 := react.NewValueFrom(0)
vInt2.Subscribe(s)

// Set action on change
vInt.OnChange(func(i interface{}) {
    fmt.Println(i)
})

// Bind another value
vInt32 := react.NewValue()
vInt32.Bind(vInt, func(v interface{}) interface{} {
    return int32(v.(int) + 1)
})

// Convert a int value to a string value
vStr := react.Convert(vInt, func(v interface{}) interface{} {
    return fmt.Sprint(v.(int) + 2)
})

// Send a value to Source
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
