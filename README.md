# chanchain

chanchain for Golang1.17-.

[See here for Golang1.18+](https://github.com/Nomango/chanchain).

## Usage

```golang
ch := make(chan int)

// Create a source
s := chanchain.NewSource(ch)

// Listen the source and get a value returned
vInt := s.Listen(context.Background())

// Convert a int Value to a int32 Value
vInt32 := chanchain.Convert(vInt, func(v interface{}) interface{} {
    return int32(v.(int))
})

// Convert a int Value to a string Value
vStr := chanchain.Convert(vInt, func(v interface{}) interface{} {
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
