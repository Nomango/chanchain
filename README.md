# go-react

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for Golang1.18+](https://github.com/Nomango/go-react).

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewSource(ch)

// Subscribe the source and get a value returned
vInt := s.Subscribe(context.Background())

// Set action on change
vInt.OnChange(func(v interface{}) {
    fmt.Println(v)
})

// Convert a int Value to a int32 Value
vInt32 := react.Convert(vInt, func(v interface{}) interface{} {
    return int32(v.(int))
})

// Convert a int Value to a string Value
vStr := react.Convert(vInt, func(v interface{}) interface{} {
    return fmt.Sprint(v)
})

// Send a value to Source
ch <- 1

fmt.Println(vInt.Load())
fmt.Println(vInt32.Load())
fmt.Println(vStr.Load())

// Output:
// 1
// 1
// 1
```
