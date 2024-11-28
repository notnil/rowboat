# RowBoat

[![Go Reference](https://pkg.go.dev/badge/github.com/notnil/rowboat.svg)](https://pkg.go.dev/github.com/notnil/rowboat)
[![Go Report Card](https://goreportcard.com/badge/github.com/notnil/rowboat)](https://goreportcard.com/report/github.com/notnil/rowboat)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)


<img src="image.png" alt="Build Status" width="250" />


RowBoat is a Go package that provides a simple and efficient way to read from and write to CSV files using Go's generics. It leverages struct tags and reflection to map CSV headers to struct fields, making it easy to work with CSV data in a type-safe manner.

## Features

- **Generic CSV Reader and Writer**: Read and write CSV data into custom structs using Go generics.
- **Struct Field Mapping**: Automatically maps CSV headers to struct fields based on field names or `csv` tags.
- **Custom Marshaling and Unmarshaling**: Support for custom types that implement `CSVMarshaler` and `CSVUnmarshaler` interfaces.
- **Field Indexing**: Control the order of fields in CSV output using `index` in struct tags.
- **Support for Basic Types**: Handles basic Go types including `string`, `int`, `float64`, `bool`, and `time.Time`.

## Installation

```bash
go get github.com/notnil/rowboat
```

## Usage

### Defining Your Struct

Define a struct that represents the CSV data. Use struct tags to specify CSV headers and indexing if needed.

```go
package main

import (
    "github.com/notnil/rowboat"
    "time"
)

type Person struct {
    Name     string    `csv:"Name"`
    Email    string    `csv:"Email"`
    Age      int       `csv:"Age"`
    JoinedAt time.Time `csv:"JoinedAt"`
}
```

### Reading CSV Data

Create a `Reader` instance and read CSV data into your struct.

```go
// main.go

package main

import (
    "fmt"
    "os"
    "github.com/notnil/rowboat"
    "slices"
)

func main() {
    // Open your CSV file
    file, err := os.Open("people.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a new Reader instance
    rb, err := rowboat.NewReader[Person](file)
    if err != nil {
        panic(err)
    }

    // Collect all records
    people := slices.Collect(rb.All())

    // Use the data
    for _, person := range people {
        fmt.Printf("%+v\n", person)
    }
}
```

### Writing CSV Data

Create a `Writer` instance and write your struct data to a CSV file.

```go
// main.go

package main

import (
    "os"

    "github.com/notnil/rowboat"
)

func main() {
    // Open a file for writing
    file, err := os.Create("people_output.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a new Writer instance
    writer, err := rowboat.NewWriter[Person](file)
    if err != nil {
        panic(err)
    }

    // Sample data
    people := []Person{
        {Name: "Alice", Email: "alice@example.com", Age: 30},
        {Name: "Bob", Email: "bob@example.com", Age: 25},
    }

    if err := writer.WriteHeader(); err != nil {
        panic(err)
    }

    // Write all records
    for _, person := range people {
        if err := writer.Write(person); err != nil {
            panic(err)
        }
    }
}
```

## Advanced Features

### Custom Unmarshaling

If you have custom types, implement the `CSVUnmarshaler` interface to define how to parse CSV strings.

```go
// point.go

package main

import (
    "fmt"
    "strconv"
    "strings"
)

type Point struct {
    X, Y float64
}

func (p *Point) UnmarshalCSV(value string) error {
    parts := strings.Split(value, ";")
    if len(parts) != 2 {
        return fmt.Errorf("invalid point format")
    }
    x, err := strconv.ParseFloat(parts[0], 64)
    if err != nil {
        return err
    }
    y, err := strconv.ParseFloat(parts[1], 64)
    if err != nil {
        return err
    }
    p.X = x
    p.Y = y
    return nil
}
```

### Custom Marshaling

For writing custom types, implement the `CSVMarshaler` interface.

```go
// point.go

func (p Point) MarshalCSV() (string, error) {
    return fmt.Sprintf("%.2f;%.2f", p.X, p.Y), nil
}
```

### Field Indexing

Control the order of fields in the CSV output using the `index` tag.

```go
type IndexedPerson struct {
    Age   int    `csv:"Age,index=0"`
    Name  string `csv:"Name,index=1"`
    Email string `csv:"Email,index=2"`
}
```

## Examples

### Reading with Filters

Use the `Filter` function to read only specific records.

```go
// main.go

package main

import (
    "fmt"
    "slices"
    "strings"

    "github.com/notnil/rowboat"
)

func main() {
    csvData := `Name,Email,Age
Alice,alice@example.com,30
Bob,bob@example.com,25
Charlie,charlie@example.com,35`

    reader := strings.NewReader(csvData)
    rb, err := rowboat.NewReader[Person](reader)
    if err != nil {
        panic(err)
    }

    adults := slices.Collect(rowboat.Filter(func(p Person) bool {
        return p.Age >= 30
    }, rb.All()))

    fmt.Println(adults)
}
```

### Writing All Records from an Iterator

```go
// main.go

package main

import (
    "iter"
    "os"
    "slices"

    "github.com/notnil/rowboat"
)

func main() {
    file, err := os.Create("people_output.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    writer, err := rowboat.NewWriter[Person](file)
    if err != nil {
        panic(err)
    }

    people := []Person{
        {Name: "Alice", Email: "alice@example.com", Age: 30},
        {Name: "Bob", Email: "bob@example.com", Age: 25},
    }

    if err := writer.WriteHeader(); err != nil {
        panic(err)
    }

    // Write all records
    if err := writer.WriteAll(slices.Values(people)); err != nil {
        panic(err)
    }
}
```

## Struct Tag Details

- **`csv:"ColumnName"`**: Specifies the CSV header name for the field.
- **`csv:"-"`**: Skips the field; it will not be read from or written to CSV.
- **`index=N`**: Sets the index (order) of the field in the CSV. Lower indexes come first.

## Custom Types Interface Definitions

```go
// reader.go

type CSVUnmarshaler interface {
    UnmarshalCSV(string) error
}
```

```go
// writer.go

type CSVMarshaler interface {
    MarshalCSV() (string, error)
}
```

## License

This project is licensed under the MIT License.
