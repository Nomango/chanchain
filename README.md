# chanchain

chanchain for Golang1.18+.

[See here for earlier version of Golang](https://github.com/Nomango/chanchain/tree/legacy).

## Usage

```golang
ch := make(chan int)

// Create a source
s := chanchain.NewSource(ch)

// Listen the source and get a value returned
vInt := s.Listen(context.Background())

// Set action on change
vInt.OnChange(func(i int) {
    fmt.Println(i)
})

// Convert a int Value to a int32 Value
vInt32 := chanchain.Convert(vInt, func(v int) int32 {
    return int32(v)
})

// Convert a int Value to a string Value
vStr := chanchain.Convert(vInt, func(v int) string {
    return fmt.Sprint(v)
})

// Send a value to Source
ch <- 1

fmt.Println(vInt32.Load())
fmt.Println(vStr.Load())

// Output:
// 1
// 1
// 1
```
