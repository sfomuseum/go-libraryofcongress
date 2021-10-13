// package walk provides methods for walking all in the records in a line-delimited JSON document.
package walk

import (
	"fmt"
	"github.com/aaronland/go-json-query"
	"io"
)

const CONTEXT_PATH string = "github.com/aaronland/go-jsonl#path"

type WalkOptions struct {
	URI           string
	Workers       int
	RecordChannel chan *WalkRecord
	ErrorChannel  chan *WalkError
	ValidateJSON  bool
	FormatJSON    bool
	QuerySet      *query.QuerySet
	IsBzip        bool
}

type WalkRecord struct {
	Path       string
	LineNumber int
	Body       []byte
}

type WalkError struct {
	Path       string
	LineNumber int
	Err        error
}

func (e *WalkError) Error() string {
	return e.String()
}

func (e *WalkError) String() string {
	return fmt.Sprintf("[%s] line %d, %v", e.Path, e.LineNumber, e.Err)
}

func IsEOFError(err error) bool {

	switch err.(type) {
	case *WalkError:

		if err.(*WalkError).Err == io.EOF {
			return true
		}

		return false
	default:
		return false
	}
}
