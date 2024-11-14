## 🚀 all: A Parallelism Tool for Go

**all** is a powerful parallelism tool for Go. It offers an easy way to run multiple goroutines concurrently and gather all the results together seamlessly.

### 📦 Installation

```bash
go get go.yuchanns.xyz/all
```

### 🔧 Usage

1. Basic Usage:

```go
package main

import (
    "fmt"
    "time"

    "go.yuchanns.xyz/all"
)

func consumer(ctx context.Context, input int32) (output string, err error) {
    // Process the input
    output = fmt.Sprintf("input: %d", input)
    return
}

func main() {
    // Create a new all executor instance
    x := all.New(context.Background(), consumer)

    // Assign a task to the executor instance
    x.Assign(1)

    // Assign another task to the executor instance
    x.Assign(2)

    // Iterate through the results
    for x.Next() {
        result := x.Each()
    }

    // Check for any errors
    err := x.Error()
}
```

2. Collect the results in a slice following the input order:

```go
results, err := all.Collect(x)
```

3. By default, the executor stops when any error occurs. You can change this behavior:

```go
all.Persist(x)
```

### 📜 License
[Apache License 2.0](LICENSE)

