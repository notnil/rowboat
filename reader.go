package rowboat

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CSVUnmarshaler is an interface for custom CSV unmarshaling
type CSVUnmarshaler interface {
	UnmarshalCSV(string) error
}

// Reader struct holds the CSV reader and mapping information
type Reader[T any] struct {
	reader   *csv.Reader
	headers  []string
	fieldMap map[int]reflect.StructField
	err      error
	current  T
}

// NewReader creates a new RowBoat reader instance
func NewReader[T any](r io.Reader) (*Reader[T], error) {
	rb := &Reader[T]{}
	rb.reader = csv.NewReader(r)

	// Read headers
	headers, err := rb.reader.Read()
	if err != nil {
		return nil, err
	}
	rb.headers = headers

	// Map CSV headers to struct fields
	if err := rb.createFieldMap(); err != nil {
		return nil, err
	}

	return rb, nil
}

// createFieldMap maps CSV headers to struct fields using struct tags
func (rb *Reader[T]) createFieldMap() error {
	rb.fieldMap = make(map[int]reflect.StructField)

	var t T
	tType := reflect.TypeOf(t)
	if tType.Kind() != reflect.Struct {
		return errors.New("generic type T must be a struct")
	}

	// First pass: collect fields and their indexes
	type fieldInfo struct {
		field reflect.StructField
		name  string
		index int
	}
	fields := make([]fieldInfo, 0, tType.NumField())
	maxIndex := -1

	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue
		}

		name := field.Name
		index := i // default index is field order
		tagParts := strings.Split(csvTag, ",")
		if len(tagParts) > 0 && tagParts[0] != "" {
			name = tagParts[0]
		}

		for _, part := range tagParts[1:] {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "index=") {
				idxStr := strings.TrimPrefix(part, "index=")
				idx, err := strconv.Atoi(idxStr)
				if err != nil {
					return fmt.Errorf("invalid index value '%s' in field '%s': %v", idxStr, field.Name, err)
				}
				index = idx
				if index > maxIndex {
					maxIndex = index
				}
			}
		}

		fields = append(fields, fieldInfo{
			field: field,
			name:  name,
			index: index,
		})
	}

	// Assign indexes to fields without explicit index
	nextIndex := maxIndex + 1
	for i := range fields {
		if fields[i].index == fields[i].field.Index[0] {
			fields[i].index = nextIndex
			nextIndex++
		}
	}

	// Sort fields by index
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].index < fields[j].index
	})

	// Map headers to fields
	headerMap := make(map[string]int)
	for i, header := range rb.headers {
		headerMap[strings.TrimSpace(header)] = i
	}

	// Create final field mapping
	for _, fi := range fields {
		if idx, ok := headerMap[fi.name]; ok {
			rb.fieldMap[idx] = fi.field
		}
	}

	return nil
}

// nextRow advances the iterator and parses the next record
func (rb *Reader[T]) nextRow() bool {
	record, err := rb.reader.Read()
	if err == io.EOF {
		return false
	}
	if err != nil {
		rb.err = err
		return false
	}

	var t T
	tValue := reflect.ValueOf(&t).Elem()

	for idx, value := range record {
		if field, ok := rb.fieldMap[idx]; ok {
			fieldValue := tValue.FieldByName(field.Name)
			if !fieldValue.CanSet() {
				continue
			}
			if err := setFieldValue(fieldValue, value); err != nil {
				rb.err = fmt.Errorf("error setting field %s: %w", field.Name, err)
				return false
			}
		}
	}
	rb.current = t
	return true
}

// setFieldValue sets the value of a struct field based on its type
func setFieldValue(field reflect.Value, value string) error {
	csvUnmarshalerType := reflect.TypeOf((*CSVUnmarshaler)(nil)).Elem()

	// Check if the field implements CSVUnmarshaler
	if field.CanInterface() && field.Type().Implements(csvUnmarshalerType) {
		unmarshaler := field.Interface().(CSVUnmarshaler)
		return unmarshaler.UnmarshalCSV(value)
	}

	// Check if the pointer to the field implements CSVUnmarshaler
	if field.CanAddr() && field.Addr().CanInterface() && field.Addr().Type().Implements(csvUnmarshalerType) {
		unmarshaler := field.Addr().Interface().(CSVUnmarshaler)
		return unmarshaler.UnmarshalCSV(value)
	}

	// Handle specific types like time.Time
	if field.Type() == reflect.TypeOf(time.Time{}) {
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))
		return nil
	}

	// Handle basic kinds
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}
	return nil
}

// All returns an iterator over all records in the CSV file.
// Each iteration returns a parsed struct of type T.
func (rb *Reader[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		// Keep reading records until we hit EOF or an error
		for rb.nextRow() {
			// Get the current record
			record := rb.current

			// Pass to yield function - if it returns false, stop iteration
			if !yield(record) {
				return
			}
		}

		// Check if we stopped due to an error
		if rb.err != nil && rb.err != io.EOF {
			// We can't return an error directly from the iterator,
			// but we can panic which will be caught by the range loop
			panic(rb.err)
		}
	}
}

// Filter returns a sequence that contains the elements
// of s for which f returns true.
func Filter[V any](f func(V) bool, s iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range s {
			if f(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}
