package query

import (
	"errors"
	"regexp"
	"strings"
)

// The separator string used to distinguish {PATH}={REGULAR_EXPRESSION} strings.
const SEP string = "="

// QueryFlags holds one or more Query instances that are created using {PATH}={REGULAR_EXPRESSION} strings.
type QueryFlags []*Query

// Return the string value of the set of Query instances. Currently returns "".
func (m *QueryFlags) String() string {
	return ""
}

// Parse a {PATH}={REGULAR_EXPRESSION} string and store it as one of a set of Query instances.
func (m *QueryFlags) Set(value string) error {

	parts := strings.Split(value, SEP)

	if len(parts) != 2 {
		return errors.New("Invalid query flag")
	}

	path := parts[0]
	str_match := parts[1]

	re, err := regexp.Compile(str_match)

	if err != nil {
		return err
	}

	q := &Query{
		Path:  path,
		Match: re,
	}

	*m = append(*m, q)
	return nil
}
