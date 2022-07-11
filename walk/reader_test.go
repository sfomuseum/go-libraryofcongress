package walk

import (
	"github.com/jeffallen/seekinghttp"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalWalkReader(t *testing.T) {

	local_tests := []string{
		"../fixtures/lcsh.sample.ndjson",
		"../fixtures/lcsh.sample.ndjson.zip",
	}

	for _, rel_path := range local_tests {

		abs_path, err := filepath.Abs(rel_path)

		if err != nil {
			t.Fatalf("Failed to derive absolute path for '%s', %v", rel_path, err)
		}

		fh, err := os.Open(abs_path)

		if err != nil {
			t.Fatalf("Failed to open '%s', %v", abs_path, err)
		}

		r := &LocalWalkReader{local: fh}

		err = r.Close()

		if err != nil {
			t.Fatalf("Failed to close reader for %s, %v", abs_path, err)
		}
	}

}

func TestRemoteWalkReader(t *testing.T) {

	remote_tests := []string{
		"https://id.loc.gov/download/lcsh.both.ndjson.zip",
	}

	for _, uri := range remote_tests {

		req := seekinghttp.New(uri)

		r := &RemoteWalkReader{remote: req}

		err := r.Close()

		if err != nil {
			t.Fatalf("Failed to close reader for %s, %v", uri, err)
		}

	}
}
