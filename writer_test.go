package rowboat_test

import (
	"bytes"
	"iter"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/notnil/rowboat"
)

func TestWriter(t *testing.T) {
	// Create test data
	people := []Person{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer

	// Create a new Writer instance
	writer, err := rowboat.NewWriter[Person](&buf)
	if err != nil {
		t.Fatalf("Failed to create Writer: %v", err)
	}

	// Write all records
	for _, person := range people {
		if err := writer.Write(person); err != nil {
			t.Fatalf("Failed to write record: %v", err)
		}
	}

	// Get the written CSV data
	csvData := buf.String()

	// Create a reader to verify the written data
	reader := strings.NewReader(csvData)
	rb, err := rowboat.NewReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	// Read back the data and compare
	results := slices.Collect(rb.All())

	if !reflect.DeepEqual(results, people) {
		t.Errorf("Written results do not match expected.\nExpected: %+v\nGot: %+v", people, results)
	}
}

func TestComplexWriter(t *testing.T) {
	// Create test data with complex types
	t1 := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	t2 := time.Date(2023, 6, 15, 9, 30, 0, 0, time.UTC)

	records := []ComplexRecord{
		{
			Name:      "John",
			CreatedAt: t1,
			Active:    true,
			Score:     98.6,
			Count:     42,
			Rate:      3.14,
			Tags:      "test;debug",
		},
		{
			Name:      "Jane",
			CreatedAt: t2,
			Active:    false,
			Score:     75.2,
			Count:     100,
			Rate:      2.718,
			Tags:      "prod;live",
		},
	}

	var buf bytes.Buffer
	writer, err := rowboat.NewWriter[ComplexRecord](&buf)
	if err != nil {
		t.Fatalf("Failed to create Writer: %v", err)
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record: %v", err)
		}
	}

	// Read back and verify
	reader := strings.NewReader(buf.String())
	rb, err := rowboat.NewReader[ComplexRecord](reader)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	results := slices.Collect(rb.All())

	if !reflect.DeepEqual(results, records) {
		t.Errorf("Written results do not match expected.\nExpected: %+v\nGot: %+v", records, results)
	}
}

func TestWriteAll(t *testing.T) {
	t1 := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	t2 := time.Date(2023, 6, 15, 9, 30, 0, 0, time.UTC)

	records := []ComplexRecord{
		{
			Name:      "John",
			CreatedAt: t1,
			Active:    true,
			Score:     98.6,
			Count:     42,
			Rate:      3.14,
			Tags:      "test;debug",
		},
		{
			Name:      "Jane",
			CreatedAt: t2,
			Active:    false,
			Score:     75.2,
			Count:     100,
			Rate:      2.718,
			Tags:      "prod;live",
		},
	}

	var buf bytes.Buffer
	writer, err := rowboat.NewWriter[ComplexRecord](&buf)
	if err != nil {
		t.Fatalf("Failed to create Writer: %v", err)
	}

	// Convert slice to iterator using iter.Seq
	recordIter := iter.Seq[ComplexRecord](func(yield func(ComplexRecord) bool) {
		for _, record := range records {
			if !yield(record) {
				return
			}
		}
	})

	if err := writer.WriteAll(recordIter); err != nil {
		t.Fatalf("Failed to write records: %v", err)
	}

	// Read back and verify
	reader := strings.NewReader(buf.String())
	rb, err := rowboat.NewReader[ComplexRecord](reader)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	results := slices.Collect(rb.All())

	if !reflect.DeepEqual(results, records) {
		t.Errorf("Written results do not match expected.\nExpected: %+v\nGot: %+v", records, results)
	}
}

func TestCustomMarshaler(t *testing.T) {
	records := []Custom{
		{
			Point: Point{X: 1.23, Y: 4.56},
		},
		{
			Point: Point{X: 7.89, Y: 0.12},
		},
	}

	var buf bytes.Buffer
	writer, err := rowboat.NewWriter[Custom](&buf)
	if err != nil {
		t.Fatalf("Failed to create Writer: %v", err)
	}

	// Convert slice to iterator using iter.Seq
	recordIter := iter.Seq[Custom](func(yield func(Custom) bool) {
		for _, record := range records {
			if !yield(record) {
				return
			}
		}
	})

	if err := writer.WriteAll(recordIter); err != nil {
		t.Fatalf("Failed to write records: %v", err)
	}

	expected := "point\n1.23;4.56\n7.89;0.12\n"
	if buf.String() != expected {
		t.Errorf("Written CSV does not match expected.\nExpected:\n%s\nGot:\n%s", expected, buf.String())
	}
}

func TestWriterWithIndexing(t *testing.T) {
	// Create test data
	people := []Person{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	// Define Person with indexing
	type IndexedPerson struct {
		Name  string `csv:"Name,index=1"`
		Email string `csv:"Email,index=2"`
		Age   int    `csv:"Age,index=0"`
	}

	// Convert people to IndexedPerson
	indexedPeople := make([]IndexedPerson, len(people))
	for i, p := range people {
		indexedPeople[i] = IndexedPerson{
			Name:  p.Name,
			Email: p.Email,
			Age:   p.Age,
		}
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer

	// Create a new Writer instance
	writer, err := rowboat.NewWriter[IndexedPerson](&buf)
	if err != nil {
		t.Fatalf("Failed to create Writer: %v", err)
	}

	// Write all records
	for _, person := range indexedPeople {
		if err := writer.Write(person); err != nil {
			t.Fatalf("Failed to write record: %v", err)
		}
	}

	// Expected CSV output
	expectedCSV := `Age,Name,Email
30,Alice,alice@example.com
25,Bob,bob@example.com
35,Charlie,charlie@example.com
`

	csvStr := buf.String()
	if csvStr != expectedCSV {
		t.Errorf("Written CSV does not match expected.\nExpected:\n%s\nGot:\n%s", expectedCSV, csvStr)
	}

	// Verify by reading back
	reader := strings.NewReader(csvStr)
	rb, err := rowboat.NewReader[IndexedPerson](reader)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	// Read back the data and compare
	results := slices.Collect(rb.All())

	// Convert back to original person for comparison
	readPeople := make([]Person, len(results))
	for i, ip := range results {
		readPeople[i] = Person{
			Name:  ip.Name,
			Email: ip.Email,
			Age:   ip.Age,
		}
	}

	if !reflect.DeepEqual(readPeople, people) {
		t.Errorf("Written and read results do not match expected.\nExpected: %+v\nGot: %+v", people, readPeople)
	}
}
