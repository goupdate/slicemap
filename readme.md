# SliceMap Library

## Overview

`SliceMap` is a Go library that provides a thread-safe, ordered mapping from keys to slices of values. It ensures that the slices are kept sorted and unique for efficient operations. This library is particularly useful for applications that require fast lookup, addition, and deletion of elements with maintained order.

## Features

- Thread-safe operations with fine-grained locking.
- Maintain sorted order of elements in the slices.
- Efficient addition and deletion of elements with binary search.
- Utility methods for checking existence, iterating keys, and values.

## Installation

To install `Slice
Map`, use `go get`:

```bash
go get -u github.com/goupdate/slicemap
```

## Usage

### Creating a new SliceMap

```go
import "github.com/goupdate/slicemap"

// Create a new SliceMap instance
sm := slicemap.NewSliceMap[int, int]()
```

### Adding elements

```go
// Add elements to the map
sm.Add(1, 10)
sm.Add(1, 20)
sm.Add(2, 15)
```

### Checking for existence

```go
// Check if an element exists
exists := sm.Exist(1, 10) // returns true
```

### Deleting elements

```go
// Delete an element
sm.Delete(1, 10)

// Delete a key and its associated slice
sm.DeleteKey(2)
```

### Iterating over elements

```go
// Iterate over all values
sm.IterateValues(func(k, v int) bool {
    fmt.Printf("Key: %d, Value: %d\n", k, v)
    return true // return false to break the iteration
})

// Iterate over keys
sm.IterateKeys(func(k int) bool {
    fmt.Println("Key:", k)
    return true
})
```


