# go-react

Not React.js for Golang!

`go-react` is a library for data binding.

[See here for earlier version of Golang](https://github.com/Nomango/go-react/tree/legacy).

## Usage

```golang
ch := make(chan int)

// Create a source
s := react.NewSource(ch)

// Create a value and subscribe the source
vInt := react.NewValue(0)
vInt.Subscribe(s)

// A source can be subscribed more than one time
// So the following code is valid
vInt2 := react.NewValue(0)
vInt2.Subscribe(s)

// Set action on change
vInt.OnChange(func(i int) {
    fmt.Println(i)
})

// Bind another value
var vInt32 react.Value[int32]
react.Bind(vInt, &vInt32, func(v int) int32 {
    return int32(v + 1)
})

// Convert a int value to a string value
vStr := react.Convert(vInt, func(v int) string {
    return fmt.Sprint(v+2)
})

// Send a value to Source
ch <- 1

fmt.Println(vInt32.Load())
fmt.Println(vStr.Load())

// Output:
// 1
// 2
// 3
```
