# chanchain

## Usage

```golang
ch := make(chan int)

// Create a Chain
c := chanchain.NewChain(func(v interface{}) interface{} {
    return fmt.Sprint(v) + "1"
})
c.Append(func(v interface{}) interface{} {
    return fmt.Sprint(v) + "2"
})

// Start a chain with a source
c.Start(ctx, chanchain.NewSource(ch))

// Create a Value
v := chanchain.NewValue(c)

// Default value of a Value is nil
fmt.Println(v.Load())

// Input
ch <- 0
time.Sleep(time.Second)

// Get latest value
fmt.Println(v.Load())

// Output:
// nil
// 012
```
