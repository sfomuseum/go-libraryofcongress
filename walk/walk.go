package walk

import (
	"context"
	"io"
)

type WalkCallbackFunction func(context.Context, []byte) error

type Walker interface {
	WalkURIs(context.Context, WalkCallbackFunction, ...string) error
	WalkFile(context.Context, WalkCallbackFunction, string) error
	WalkZipFile(context.Context, WalkCallbackFunction, string) error
	WalkReader(context.Context, WalkCallbackFunction, io.Reader) error
}
