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

// CSVMarshaler is an interface for custom CSV marshaling
type CSVMarshaler interface {
	MarshalCSV() (string, error)
}

// fieldInfo contains information about a struct field for CSV writing
type fieldInfo struct {
	Index int
	Name  string
	Field reflect.StructField
}

// Writer struct holds the CSV writer and mapping information
type Writer[T any] struct {
	writer          *csv.Writer
	fields          []fieldInfo
	writeHeaderOnce bool
}

// NewWriter creates a new RowBoat writer instance
func NewWriter[T any](w io.Writer) (*Writer[T], error) {
	rw := &Writer[T]{}
	rw.writer = csv.NewWriter(w)

	// Analyze the struct fields
	if err := rw.createFieldInfo(); err != nil {
		return nil, err
	}

	return rw, nil
}

// createFieldInfo extracts information about struct fields, including indexes
func (rw *Writer[T]) createFieldInfo() error {
	var t T
	tType := reflect.TypeOf(t)
	if tType.Kind() != reflect.Struct {
		return errors.New("generic type T must be a struct")
	}

	fields := make([]fieldInfo, 0, tType.NumField())
	maxIndex := -1

	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue // skip field
		}

		name := field.Name
		index := i // default index is the field order
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
			Index: index,
			Name:  name,
			Field: field,
		})
	}

	// Assign indexes to fields without an explicit index, starting from maxIndex+1
	nextIndex := maxIndex + 1
	for i := range fields {
		if fields[i].Index == fields[i].Field.Index[0] { // Field's default index
			fields[i].Index = nextIndex
			nextIndex++
		}
	}

	// Sort the fields based on the index
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Index < fields[j].Index
	})

	rw.fields = fields
	return nil
}

// Write writes a single record to the CSV writer
func (rw *Writer[T]) Write(record T) error {
	// Write header if not written yet
	if !rw.writeHeaderOnce {
		headers := make([]string, len(rw.fields))
		for i, fi := range rw.fields {
			headers[i] = fi.Name
		}
		if err := rw.writer.Write(headers); err != nil {
			return err
		}
		rw.writer.Flush() // Ensure headers are written
		rw.writeHeaderOnce = true
	}

	recordValues := make([]string, len(rw.fields))
	v := reflect.ValueOf(record)
	for i, fi := range rw.fields {
		fieldValue := v.FieldByName(fi.Field.Name)
		strValue, err := getFieldStringValue(fieldValue)
		if err != nil {
			return fmt.Errorf("error marshaling field %s: %w", fi.Field.Name, err)
		}
		recordValues[i] = strValue
	}

	if err := rw.writer.Write(recordValues); err != nil {
		return err
	}
	rw.writer.Flush()
	return nil
}

// getFieldStringValue converts a struct field value to string for CSV
func getFieldStringValue(field reflect.Value) (string, error) {
	csvMarshalerType := reflect.TypeOf((*CSVMarshaler)(nil)).Elem()

	// Check if the field implements CSVMarshaler
	if field.CanInterface() && field.Type().Implements(csvMarshalerType) {
		marshaler := field.Interface().(CSVMarshaler)
		return marshaler.MarshalCSV()
	}

	// Check if the pointer to the field implements CSVMarshaler
	if field.CanAddr() && field.Addr().CanInterface() && field.Addr().Type().Implements(csvMarshalerType) {
		marshaler := field.Addr().Interface().(CSVMarshaler)
		return marshaler.MarshalCSV()
	}

	// Handle specific types like time.Time
	if field.Type() == reflect.TypeOf(time.Time{}) {
		t := field.Interface().(time.Time)
		return t.Format(time.RFC3339), nil
	}

	// Handle basic kinds
	switch field.Kind() {
	case reflect.String:
		return field.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(field.Bool()), nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", field.Type())
	}
}

// WriteAll writes multiple records from an iterator
func (rw *Writer[T]) WriteAll(records iter.Seq[T]) error {
	var err error
	records(func(record T) bool {
		if err = rw.Write(record); err != nil {
			return false
		}
		return true
	})
	return err
}
