package qstring

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// parseOperator parses a leading logical operator out of the provided string
func parseOperator(s string) string {
	switch s[0] {
	case 60: // "<"
		switch s[1] {
		case 61: // "="
			return "<="
		default:
			return "<"
		}
	case 62: // ">"
		switch s[1] {
		case 61: // "="
			return ">="
		default:
			return ">"
		}
	default:
		// no operator found, default to "="
		return "="
	}
}

// ComparativeTime is a field that can be used for specifying a query parameter
// which includes a conditional operator and a timestamp
type ComparativeTime struct {
	Operator string
	Time     time.Time
}

// NewComparativeTime returns a new ComparativeTime instance with a default
// operator of "="
func NewComparativeTime() *ComparativeTime {
	return &ComparativeTime{Operator: "="}
}

// Parse is used to parse a query string into a ComparativeTime instance
func (c *ComparativeTime) Parse(query string) error {
	if len(query) <= 2 {
		return errors.New("qstring: Invalid Timestamp Query")
	}

	c.Operator = parseOperator(query)

	// if no operator was provided and we defaulted to an equality operator
	if !strings.HasPrefix(query, c.Operator) {
		query = fmt.Sprintf("=%s", query)
	}

	var err error
	c.Time, err = time.Parse(time.RFC3339, query[len(c.Operator):])
	if err != nil {
		return err
	}

	return nil
}

// String returns this ComparativeTime instance in the form of the query
// parameter that it came in on
func (c ComparativeTime) String() string {
	return fmt.Sprintf("%s%s", c.Operator, c.Time.Format(time.RFC3339))
}
