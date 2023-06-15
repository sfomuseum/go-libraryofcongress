// Package walk provides interfaces and methods for walking Library of Congress (LoC) data files.
package walk

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// type WalkCallbackFunction defines a user-specified callback function for processing a LoC data file.
type WalkCallbackFunction func(context.Context, []byte) error

// type Walker defines an interface for iterating (walking) LoC data files from a variety or sources.
type Walker interface {
	// WalkURIs iterates (walks) LoC data files from one or more URIs.
	WalkURIs(context.Context, WalkCallbackFunction, ...string) error
	// WalkFile iterates (walks) a LoC data file on disk.
	WalkFile(context.Context, WalkCallbackFunction, string) error
	// WalkZipFile iterates (walks) a LoC zip-compressed data file on disk.
	WalkZipFile(context.Context, WalkCallbackFunction, string) error
	// WalkZipFile iterates (walks) LoC data from an `io.Reader` instance.
	WalkReader(context.Context, WalkCallbackFunction, io.Reader) error
}

// type WalkerInitializeFunc is a function used to initialize an implementation of the `Walker` interface.
type WalkerInitializeFunc func(ctx context.Context, uri string) (Walker, error)

// walkers is a `aaronland/go-roster.Roster` instance used to maintain a list of registered `WalkerInitializeFunc` initialization functions.
var walkers roster.Roster

// ensureWalkerRoster() ensures that a `aaronland/go-roster.Roster` instance used to maintain a list of registered `WalkerInitializeFunc`
// initialization functions is present
func ensureWalkerRoster() error {

	if walkers == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create new roster, %w", err)
		}

		walkers = r
	}

	return nil
}

// RegisterWalker() associates 'scheme' with 'init_func' in an internal list of avilable `Walker` implementations.
func RegisterWalker(ctx context.Context, scheme string, f WalkerInitializeFunc) error {

	err := ensureWalkerRoster()

	if err != nil {
		return fmt.Errorf("Failed to ensure roster, %w", err)
	}

	return walkers.Register(ctx, scheme, f)
}

// Schemes() returns the list of schemes that have been "registered".
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureWalkerRoster()

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

// NewWalker() returns a new `Walker` instance derived from 'uri'. The semantics of and requirements for
// 'uri' as specific to the package implementing the interface.
func NewWalker(ctx context.Context, uri string) (Walker, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	i, err := walkers.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive walker for '%s', %w", scheme, err)
	}

	f := i.(WalkerInitializeFunc)
	return f(ctx, uri)
}
