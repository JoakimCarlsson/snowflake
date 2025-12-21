# Snowflake

A lightweight, thread-safe implementation of Twitter's Snowflake ID generator in Go.

## Features

- Generates unique 64-bit IDs
- Thread-safe using atomic operations
- Supports up to 1024 machines (10-bit machine ID)
- 4096 unique IDs per millisecond per machine (12-bit sequence)
- Custom epoch starting January 1, 2025
- Zero dependencies

## Installation

```bash
go get github.com/joakimcarlsson/snowflake
```

## Usage

### Generate an ID

```go
package main

import (
    "fmt"
    "github.com/joakimcarlsson/snowflake"
)

func main() {
    id := snowflake.Generate()
    fmt.Println(id) // e.g., 123456789012345678
}
```

### Parse an ID

```go
timestamp, machineID, sequence := snowflake.Parse(id)
fmt.Printf("Timestamp: %d, Machine: %d, Sequence: %d\n", timestamp, machineID, sequence)
```

## Configuration

Set the `MACHINE_ID` environment variable to specify the machine ID (0-1023). Defaults to 1 if not set.

```bash
export MACHINE_ID=42
```

## ID Structure

```
 64-bit ID
 ├── 41 bits: timestamp (milliseconds since epoch)
 ├── 10 bits: machine ID (0-1023)
 └── 12 bits: sequence number (0-4095)
```

## License

MIT License - see [LICENSE](LICENSE) for details.
