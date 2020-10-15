# gcmap
Thread safe map with record TTL capability

# Installation

```bash
~ $ go get -u github.com/hasansino/gcmap
```

# Example usage
```go
func main() {
	
    // create new instance
    st := gcmap.NewStorage(
        WithGCInterval(time.Minute),
        WithEntryTTL(time.Second * 30),
    )
    
    // create new key with value
    st.Store("key", "value")
    
    // retrieve value by key
    value, found := st.Load("key")
    if found {
    	fmt.Printf("Value is %v", value)
    }
    
    // update existing key
    st.StoreOrUpdate("key", "new_value", func(old, new interface{}) interface{} {
    	return old.(string) + new.(string) 
    })
    
    // delete key
    st.Delete("key")
}
```

# Using StoreOrUpdate
```go
type ComplexCounter struct {
	CounterA int64
	CounterB int64
}

func main() {
	
    st := gcmap.NewStorage()
    st.Store("key", ComplexCounter{CounterA: 5})
    
    // update existing complex structure
    
    st.StoreOrUpdate("key", ComplexCounter{CounterB: 10}, func(old, new interface{}) interface{} {
    	counter := old.(ComplexCounter)
    	counter.B = new.(ComplexCounter).B
    	return counter
    })
}
```