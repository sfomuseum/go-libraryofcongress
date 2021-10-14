// package walk provides interfaces and methods for walking Library of Congress data files.
package walk

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"io"
	"net/url"
	"sort"
	"strings"
)

type WalkCallbackFunction func(context.Context, []byte) error

type Walker interface {
	WalkURIs(context.Context, WalkCallbackFunction, ...string) error
	WalkFile(context.Context, WalkCallbackFunction, string) error
	WalkZipFile(context.Context, WalkCallbackFunction, string) error
	WalkReader(context.Context, WalkCallbackFunction, io.Reader) error
}

type WalkerInitializeFunc func(ctx context.Context, uri string) (Walker, error)

var walkers roster.Roster

func ensureSpatialRoster() error {

	if walkers == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		walkers = r
	}

	return nil
}

func RegisterWalker(ctx context.Context, scheme string, f WalkerInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return err
	}

	return walkers.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range walkers.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewWalker(ctx context.Context, uri string) (Walker, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := walkers.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(WalkerInitializeFunc)
	return f(ctx, uri)
}
