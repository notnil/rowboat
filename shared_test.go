package rowboat_test

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Person struct {
	Name  string `csv:"Name"`
	Email string `csv:"Email"`
	Age   int    `csv:"Age"`
}

type ComplexRecord struct {
	Name      string    `csv:"name"`
	CreatedAt time.Time `csv:"created_at"`
	Active    bool      `csv:"active"`
	Score     float64   `csv:"score"`
	Count     int       `csv:"count"`
	Rate      float32   `csv:"rate"`
	Tags      string    `csv:"tags"`
}

type Point struct {
	X float64 `csv:"x"`
	Y float64 `csv:"y"`
}

func (p *Point) UnmarshalCSV(value string) error {
	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return fmt.Errorf("invalid point format: %s", value)
	}
	var err error
	p.X, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return err
	}
	p.Y, err = strconv.ParseFloat(parts[1], 64)
	return err
}

func (p Point) MarshalCSV() (string, error) {
	return fmt.Sprintf("%.2f;%.2f", p.X, p.Y), nil
}

type Custom struct {
	Point Point `csv:"point"`
}
