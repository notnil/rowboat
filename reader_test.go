package rowboat_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/notnil/rowboat"
)

func TestRowBoat(t *testing.T) {
	// CSV data as a string
	csvData := `Name,Email,Age
Alice,alice@example.com,30
Bob,bob@example.com,25
Charlie,charlie@example.com,35`

	// Create a reader from the CSV data string
	reader := strings.NewReader(csvData)

	// Create a new RowBoat instance
	rb, err := rowboat.NewReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create RowBoat: %v", err)
	}

	// Expected results
	expected := []Person{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}
	results := slices.Collect(rowboat.Filter(func(p Person) bool {
		return p.Age > 25
	}, rb.All()))

	// Compare results
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Parsed results do not match expected.\nExpected: %+v\nGot: %+v", expected, results)
	}
}

func TestExtraneousColumns(t *testing.T) {
	// CSV data with extra columns that aren't in the struct
	csvData := `Name,Email,Age,ExtraCol1,ExtraCol2
Alice,alice@example.com,30,unused1,unused2
Bob,bob@example.com,25,unused3,unused4`

	reader := strings.NewReader(csvData)

	// Create a new RowBoat instance
	rb, err := rowboat.NewReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create RowBoat: %v", err)
	}

	// Expected results - should ignore the extra columns
	expected := []Person{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
	}

	results := slices.Collect(rb.All())

	// Compare results
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Parsed results do not match expected.\nExpected: %+v\nGot: %+v", expected, results)
	}
}

func TestBlankRows(t *testing.T) {
	// CSV data with blank rows
	csvData := `Name,Email,Age

Alice,alice@example.com,30

Bob,bob@example.com,25
`
	reader := strings.NewReader(csvData)

	// Create a new RowBoat instance
	rb, err := rowboat.NewReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create RowBoat: %v", err)
	}

	// Expected results - should skip blank rows
	expected := []Person{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
	}

	results := slices.Collect(rb.All())

	// Compare results
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Parsed results do not match expected.\nExpected: %+v\nGot: %+v", expected, results)
	}
}

func TestComplexTypes(t *testing.T) {
	// CSV data with various types
	csvData := `name,created_at,active,score,count,rate,tags
John,2023-01-02T15:04:05Z,true,98.6,42,3.14,test;debug
Jane,2023-06-15T09:30:00Z,false,75.2,100,2.718,prod;live`

	reader := strings.NewReader(csvData)

	rb, err := rowboat.NewReader[ComplexRecord](reader)
	if err != nil {
		t.Fatalf("Failed to create RowBoat: %v", err)
	}

	// Parse expected time values
	t1, _ := time.Parse(time.RFC3339, "2023-01-02T15:04:05Z")
	t2, _ := time.Parse(time.RFC3339, "2023-06-15T09:30:00Z")

	expected := []ComplexRecord{
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

	results := slices.Collect(rb.All())

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Parsed results do not match expected.\nExpected: %+v\nGot: %+v", expected, results)
	}
}

func TestCustomUnmarshaler(t *testing.T) {
	csvData := `point
1;2
3;4`
	reader := strings.NewReader(csvData)

	rb, err := rowboat.NewReader[Custom](reader)
	if err != nil {
		t.Fatalf("Failed to create RowBoat: %v", err)
	}

	results := slices.Collect(rb.All())

	expected := []Custom{
		{Point: Point{X: 1, Y: 2}},
		{Point: Point{X: 3, Y: 4}},
	}

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Parsed results do not match expected.\nExpected: %+v\nGot: %+v", expected, results)
	}
}
