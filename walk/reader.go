package walk

import (
	"os"

	"github.com/jeffallen/seekinghttp"
)

// WalkReader is an interface which implements the `io.Reader`, `io.ReaderAt` and `io.Closer` interface
// for reading Library of Congress data files. This provides a common interface for reading local and remote
// data files regardless of whether or not they are compressed.
type WalkReader interface {
	// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch space during the call. If some data is available but not len(p) bytes, Read conventionally returns what is available instead of waiting for more.
	Read(p []byte) (int, error)
	// ReadAt reads len(buf) bytes into buf starting at offset off.
	ReadAt([]byte, int64) (int, error)
	// Close closes any underlying file handles. It is implementation specific.
	Close() error
}

// type LocalWalkReader implements the `WalkReader` interface for files on a local disk.
type LocalWalkReader struct {
	WalkReader
	// local is the underlying `os.File` instance used to read data.
	local *os.File
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch space during the call. If some data is available but not len(p) bytes, Read conventionally returns what is available instead of waiting for more.
func (r *LocalWalkReader) Read(p []byte) (int, error) {
	return r.local.Read(p)
}

// ReadAt reads len(buf) bytes into buf starting at offset off.
func (r *LocalWalkReader) ReadAt(p []byte, off int64) (int, error) {
	return r.local.ReadAt(p, off)
}

// Close closes the underlying `os.File` instance for 'r'.
func (r *LocalWalkReader) Close() error {
	return r.local.Close()
}

// https://blog.gopheracademy.com/advent-2017/seekable-http/
// https://github.com/jeffallen/seekinghttp

// type RemoteWalkReader implements the `WalkReader` interface for files on a remote web server.
type RemoteWalkReader struct {
	WalkReader
	// remote is the underlying `seekinghttp.SeekingHTTP` instance used to read data.
	remote *seekinghttp.SeekingHTTP
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch space during the call. If some data is available but not len(p) bytes, Read conventionally returns what is available instead of waiting for more.
func (r *RemoteWalkReader) Read(p []byte) (int, error) {
	return r.remote.Read(p)
}

// ReadAt reads len(buf) bytes into buf starting at offset off.
func (r *RemoteWalkReader) ReadAt(p []byte, off int64) (int, error) {
	return r.remote.ReadAt(p, off)
}

// Close is a no-op.
func (r *RemoteWalkReader) Close() error {
	return nil
}
